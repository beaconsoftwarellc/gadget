package sqs

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const (
	serviceAttributeName = "service"
	methodAttributeName  = "method"
	awsTraceHeaderName   = "AWSTraceHeader"
	maxWaitTime          = 20 * time.Second
	maxMessageCount      = 10
)

// VisibilityTimeout should be used to timeout the context for messages
// 		- Functions that handle messages MUST respect the timeout
// You can extend the timeout by calling 'ChangeMessageVisibility' and setting it
//		higher than what it was. This can be done by the worker currently holding
//		the message
// Best Practice is Queue per Producer

const (
	// RegionUSEast1 is the AWS Region located in N. Virginia, USA
	RegionUSEast1 = "us-east-1"
	// LocalRegion is for testing locally with a docker instance running elasticmq
	LocalRegion = "local"
)

// New SQS instance located at the passed URL
func New(region string, queueLocator *url.URL) messagequeue.MessageQueue {
	return &sdk{
		region:   region,
		queueUrl: queueLocator,
	}
}

type sdk struct {
	region   string
	queueUrl *url.URL
	api      API
}

func (mq *sdk) API(context context.Context) (API, error) {
	if nil != mq.api {
		return mq.api, nil
	}
	var (
		err error
	)
	options := []func(*config.LoadOptions) error{config.WithRegion(mq.region)}
	if mq.region == LocalRegion {
		er := NewEndpointResolver(LocalRegion, fmt.Sprintf("http://%s",
			mq.queueUrl.Host))
		options = append(options, config.WithEndpointResolverWithOptions(er))
	}
	cfg, err := config.LoadDefaultConfig(context, options...)
	if nil == err {
		mq.api = sqs.NewFromConfig(cfg)
	}
	return mq.api, err
}

func setFailed(messages []types.SendMessageBatchRequestEntry,
	results map[string]*messagequeue.EnqueueMessageResult, errString string) {
	for _, message := range messages {
		r, ok := results[aws.ToString(message.Id)]
		if ok {
			r.Success = false
			r.SenderFault = false
			r.Error = errString
		}
	}
}

func (mq *sdk) enqueueBatch(ctx context.Context,
	messages []types.SendMessageBatchRequestEntry,
	results map[string]*messagequeue.EnqueueMessageResult) {
	if len(messages) == 0 {
		return
	}
	var (
		api  API
		err  error
		smbi *sqs.SendMessageBatchInput
		smbo *sqs.SendMessageBatchOutput
	)
	api, err = mq.API(ctx)
	if nil == err {
		smbi = &sqs.SendMessageBatchInput{
			Entries:  messages,
			QueueUrl: aws.String(mq.queueUrl.String()),
		}
		smbo, err = api.SendMessageBatch(ctx, smbi)
	}
	if err != nil {
		setFailed(messages, results, err.Error())
		return
	}
	for _, message := range smbo.Successful {
		r, ok := results[aws.ToString(message.Id)]
		if ok {
			r.ID = aws.ToString(message.MessageId)
			r.Success = true
		} else {
			log.Warnf("message ID '%s' returned from SQS is unknown",
				aws.ToString(message.Id))
		}
	}
	for _, failed := range smbo.Failed {
		r, ok := results[aws.ToString(failed.Id)]
		if ok {
			r.Success = false
			r.SenderFault = failed.SenderFault
			r.Error = fmt.Sprintf("%s: %s",
				aws.ToString(failed.Code), aws.ToString(failed.Message))
		}
	}
}

func (mq *sdk) EnqueueBatch(ctx context.Context, messages []*messagequeue.Message) (
	[]*messagequeue.EnqueueMessageResult, error) {
	if len(messages) == 0 {
		return nil, nil
	}
	var (
		err       error
		output    = make([]*messagequeue.EnqueueMessageResult, len(messages))
		resultMap = make(map[string]*messagequeue.EnqueueMessageResult)
		batch     = make([]types.SendMessageBatchRequestEntry, 0, maxMessageCount)
	)

	var smbre *types.SendMessageBatchRequestEntry
	for i, msg := range messages {
		if i != 0 && i%maxMessageCount == 0 {
			mq.enqueueBatch(ctx, batch, resultMap)
			batch = make([]types.SendMessageBatchRequestEntry, 0, maxMessageCount)
		}
		smbre, err = sendMessageBatchRequestEntryFromMessage(msg)
		output[i] = &messagequeue.EnqueueMessageResult{Message: msg}
		resultMap[*smbre.Id] = output[i]
		if nil != log.Error(err) {
			output[i].SenderFault = true
			output[i].Error = err.Error()
			continue
		}
		batch = append(batch, *smbre)
	}
	// enqueue any remaining
	mq.enqueueBatch(ctx, batch, resultMap)
	return output, nil
}

func (mq *sdk) Dequeue(ctx context.Context, count int, wait time.Duration) (
	[]*messagequeue.Message, error) {
	var (
		api      API
		err      error
		rmi      = &sqs.ReceiveMessageInput{}
		rmo      *sqs.ReceiveMessageOutput
		messages []*messagequeue.Message
	)
	api, err = mq.API(ctx)
	if nil != err {
		return nil, err
	}
	if wait < 0 {
		wait = 0
	}
	if wait > maxWaitTime {
		wait = maxWaitTime
	}
	if count < 1 {
		count = 1
	}
	if count > maxMessageCount {
		count = maxMessageCount
	}
	// We should set this here and include the timeout as a deadline on the
	// message, we can expose 'ExtendVisibilityTimeout' methods so that it
	// can be extended (up to 12 hours from receipt) as the message is processed.
	// You can provide the VisibilityTimeout parameter in your request.
	// The parameter is applied to the messages that Amazon SQS returns in the
	// response. If you don't include the parameter, the overall visibility
	// timeout for the queue is used for the returned messages.
	// rmi.SetVisibilityTimeout()
	rmi.QueueUrl = aws.String(mq.queueUrl.String())
	rmi.MaxNumberOfMessages = int32(count)
	rmi.WaitTimeSeconds = int32(wait.Seconds())
	rmo, err = api.ReceiveMessage(ctx, rmi)
	if nil != err {
		return nil, err
	}
	for _, m := range rmo.Messages {
		messages = append(messages, convert(&m))
	}
	return messages, nil
}

func (mq *sdk) Delete(ctx context.Context, msg *messagequeue.Message) error {
	var (
		api API
		err error
		dmi = &sqs.DeleteMessageInput{}
	)
	api, err = mq.API(ctx)
	if nil != err {
		return err
	}
	dmi.QueueUrl = aws.String(mq.queueUrl.String())
	dmi.ReceiptHandle = aws.String(msg.External)
	_, err = api.DeleteMessage(ctx, dmi)
	return err
}
