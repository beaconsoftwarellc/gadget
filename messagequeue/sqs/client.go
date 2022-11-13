package sqs

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const (
	serviceAttributeName          = "service"
	methodAttributeName           = "method"
	awsTraceHeaderName            = "AWSTraceHeader"
	all                    string = ".*"
	maxWaitTimeSeconds            = 20
	maxMessageDequeueCount        = 10
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
	Enqueue(m *messagequeue.Message) (*messagequeue.Message, error)
	Dequeue(count, waitSeconds int64) ([]*messagequeue.Message, error)
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

func (mq *sdk) Dequeue(count, waitTimeSeconds int64) ([]*messagequeue.Message, error) {
	var (
		client        *sqs.SQS
		err           error
		rmi           = &sqs.ReceiveMessageInput{}
		allAttributes = all
		rmo           *sqs.ReceiveMessageOutput
		messages      []*messagequeue.Message
	)
	client, err = mq.client()
	if nil != err {
		return nil, err
	}
	if waitTimeSeconds < 0 {
		waitTimeSeconds = 0
	}
	if waitTimeSeconds > maxWaitTimeSeconds {
		waitTimeSeconds = maxWaitTimeSeconds
	}
	if count < 0 {
		count = 1
	}
	if count > maxMessageDequeueCount {
		count = maxMessageDequeueCount
	}
	rmi.SetQueueUrl(mq.queueUrl.String())
	rmi.SetMaxNumberOfMessages(count)
	rmi.SetMessageAttributeNames([]*string{&allAttributes})
	rmi.SetWaitTimeSeconds(waitTimeSeconds)
	rmo, err = client.ReceiveMessage(rmi)
	if nil != err {
		return nil, err
	}
	for _, m := range rmo.Messages {
		&message{SendMessageInput: m}
		messages = append(m, messages)
	}
	return messages, nil
}
