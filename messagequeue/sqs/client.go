package sqs

import (
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const (
	serviceAttributeName   = "service"
	methodAttributeName    = "method"
	awsTraceHeaderName     = "AWSTraceHeader"
	maxWaitTime            = 20 * time.Second
	maxMessageDequeueCount = 10
)

// VisibilityTimeout should be used to timeout the context for messages
// 		- Functions that handle messages MUST respect the timeout
// You can extend the timeout by calling 'ChangeMessageVisibility' and setting it
//		higher than what it was. This can be done by the worker currently holding
//		the message
// Best Practice is Queue per Producer

// SQS interface for sending and receiving messages from a simple queueing service
// instance.
type SQS interface {
	// Enqueue the passed message
	Enqueue(m *messagequeue.Message) error
	// Enqueue all the passed messages as a batch
	EnqueueBatch(messages []*messagequeue.Message) error
	// Dequeue up to the passed count of messages waiting up to the passed
	// duration
	Dequeue(count int, wait time.Duration) ([]*messagequeue.Message, error)
	// Delete the passed message from the queue so that it is not processed by
	// other workers
	Delete(*messagequeue.Message) error
}

// New SQS instance located at the passed URL
func New(queueLocator *url.URL) SQS {
	return &sdk{
		queueUrl: queueLocator,
	}
}

type sdk struct {
	queueUrl *url.URL
	service  *sqs.SQS
}

func (mq *sdk) client() (*sqs.SQS, error) {
	if nil != mq.service {
		return mq.service, nil
	}
	var (
		err  error
		sess *session.Session
	)
	sess, err = session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if nil == err {
		mq.service = sqs.New(sess)
	}
	return mq.service, err
}

func (mq *sdk) Enqueue(msg *messagequeue.Message) error {
	var (
		client *sqs.SQS
		err    error
		smin   *sqs.SendMessageInput
		smout  *sqs.SendMessageOutput
	)
	client, err = mq.client()
	if nil != err {
		return err
	}
	smin, err = sendMessageInputFromMessage(msg)
	if nil != err {
		return err
	}

	if err = smin.Validate(); nil != err {
		return err
	}
	smout, err = client.SendMessage(smin)
	if nil == err {
		msg.ID = *smout.MessageId
	}
	return err
}

func (mq *sdk) EnqueueBatch(messages []*messagequeue.Message) error {
	if len(messages) == 0 {
		return nil
	}
	if len(messages) == 1 {
		return mq.Enqueue(messages[0])
	}
	var (
		client *sqs.SQS
		err    error
		smbi   *sqs.SendMessageBatchInput
		smbo   *sqs.SendMessageBatchOutput
	)
	client, err = mq.client()
	if nil != err {
		return err
	}
	smbi = &sqs.SendMessageBatchInput{
		Entries: make([]*sqs.SendMessageBatchRequestEntry, 0, len(messages)),
	}
	var smbre *sqs.SendMessageBatchRequestEntry
	for _, msg := range messages {
		smbre, err = sendMessageBatchRequestEntryFromMessage(msg)
		if nil != log.Error(err) {
			continue
		}
		if err = smbre.Validate(); nil != log.Error(err) {
			continue
		}
		smbi.Entries = append(smbi.Entries, smbre)
	}
	if len(smbi.Entries) == 0 {
		return errors.New("all messages were invalid")
	}
	smbi.SetQueueUrl(mq.queueUrl.String())
	if smbo, err = client.SendMessageBatch(smbi); nil != err {
		return err
	}
	// we can iterate through response.Success and response.Failed and handle
	// the specific cases if we need to later. For now just log the failures
	log.Infof("succeeded enqueueing %d of %d messages", len(smbo.Successful),
		len(messages))
	for _, failure := range smbo.Failed {
		log.Warnf("message %s failed (SenderFault: %v) with code %s: %s ",
			*failure.Id, failure.SenderFault, failure.Code, *failure.Message)
	}
	return nil
}

func (mq *sdk) Dequeue(count int, wait time.Duration) ([]*messagequeue.Message, error) {
	var (
		client   *sqs.SQS
		err      error
		rmi      = &sqs.ReceiveMessageInput{}
		rmo      *sqs.ReceiveMessageOutput
		messages []*messagequeue.Message
	)
	client, err = mq.client()
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
	if count > maxMessageDequeueCount {
		count = maxMessageDequeueCount
	}
	// We should set this here and include the timeout as a deadline on the
	// message, we can expose 'ExtendVisibilityTimeout' methods so that it
	// can be extended (up to 12 hours from receipt) as the message is processed.
	// You can provide the VisibilityTimeout parameter in your request.
	// The parameter is applied to the messages that Amazon SQS returns in the
	// response. If you don't include the parameter, the overall visibility
	// timeout for the queue is used for the returned messages.
	// rmi.SetVisibilityTimeout()
	rmi.SetQueueUrl(mq.queueUrl.String())
	rmi.SetMaxNumberOfMessages(int64(count))
	rmi.SetWaitTimeSeconds(int64(wait.Seconds()))
	rmo, err = client.ReceiveMessage(rmi)
	if nil != err {
		return nil, err
	}
	for _, m := range rmo.Messages {
		messages = append(messages, convert(m))
	}
	return messages, nil
}

func (mq *sdk) Delete(msg *messagequeue.Message) error {
	var (
		client *sqs.SQS
		err    error
		dmi    = &sqs.DeleteMessageInput{}
	)
	client, err = mq.client()
	if nil != err {
		return err
	}
	dmi.SetQueueUrl(mq.queueUrl.String())
	dmi.SetReceiptHandle(msg.External)
	_, err = client.DeleteMessage(dmi)
	return err
}
