package sqs

import (
	"errors"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const (
	serviceAttributeName = "service"
	methodAttributeName  = "method"
)

// VisibilityTimeout should be used to timeout the context for messages
// 		- Functions that handle messages MUST respect the timeout
// You can extend the timeout by calling 'ChangeMessageVisibility' and setting it
//		higher than what it was. This can be done by the worker currently holding
//		the message
// Best Practice is Queue per Producer
// Message Group ID guarantee's ORDERING
// To  avoid processing duplicate messages in a system with multiple producers
// 		and consumers where throughput and latency are more important than
//		ordering, the producer should generate a unique message group ID for
//		each message.

type SQS interface {
	SendMessage(message *messagequeue.Message) error
}

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

func (mq *sdk) SendMessage(message *messagequeue.Message) error {
	var err error
	client, err := mq.client()
	if nil != err {
		return err
	}
	sqsMessage, err := convert(message)
	mq.service.SendMessage(&sqs.SendMessageInput{
		DelaySeconds:      &delay,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{},
		MessageBody:       &body,
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// MessageDeduplicationId:  "",
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// MessageGroupId:          "",
		// the only supported message system attribute is AWSTraceHeader.
		// Its type must be String and its value must be a correctly formatted X-Ray
		// trace header string.
		MessageSystemAttributes: map[string]*sqs.MessageSystemAttributeValue{},
		QueueUrl:                &mq.queueUrl,
	})
	return err
}

func newSQSMessage(message *messagequeue.Message) (*sqsMessage, error) {
	var err error
	delay := int64(message.Delay.Seconds())
	sqsMessage := &sqsMessage{
		SendMessageInput: &sqs.SendMessageInput{
			DelaySeconds:            &delay,
			MessageAttributes:       map[string]*sqs.MessageAttributeValue{},
			MessageSystemAttributes: map[string]*sqs.MessageSystemAttributeValue{},
		},
	}
	for name, value := range message.Attributes {
		err = sqsMessage.SetAttribute(name, value)
	}
}

type sqsMessage struct {
	*sqs.SendMessageInput
}

func (m *sqsMessage) AsMessage() *messagequeue.Message {
	message := &messagequeue.Message{}
}

func (m *sqsMessage) verify(queueUrl *url.URL) (*sqs.SendMessageInput, error) {
	// make sure that the body and the queue url are set
	var err error
	qurl := queueUrl.String()
	m.SendMessageInput.QueueUrl = &qurl
	if stringutil.IsWhiteSpace(*m.MessageBody) {
		err = errors.New("message body is a required field")
	}
	return m.SendMessageInput, err
}

func (m *sqsMessage) setDelay(duration time.Duration) error {
	return nil
}

func (m *sqsMessage) setAttribute(name, value string) error {
	/*
		Name – The message attribute name can contain the following characters:
			A-Z
			a-z
			0-9
			underscore (_)
			hyphen (-)
			period (.)
		The following restrictions apply:
			- Can be up to 256 characters long
			- Can't start with AWS. or Amazon. (or any casing variations)
			- Is case-sensitive
			- Must be unique among all attribute names for the message
			- Must not start or end with a period
			- Must not have periods in a sequence
	*/

	/*
		Value – The message attribute value.
			For String data types, the attribute values has the same
			restrictions as the message body.
	*/

	return nil
}

func (m *sqsMessage) getAttribute(name, defaultValue string) string {
	value := defaultValue
	attribute, ok := m.MessageAttributes[name]
	if ok {
		value = attribute.GoString()
	}
	return value
}

func (m *sqsMessage) SetBody(body string) error {
	return nil
}

func (m *sqsMessage) verifyBody(body string) error {
	// The minimum size is one character. The maximum size is 256 KB.
	//
	// A message can include only XML, JSON, and unformatted text. The following
	// Unicode characters are allowed:
	//
	// #x9 | #xA | #xD | #x20 to #xD7FF | #xE000 to #xFFFD | #x10000 to #x10FFFF
	//
	// Any characters not included in this list will be rejected. For more information,
	// see the W3C specification for characters (http://www.w3.org/TR/REC-xml/#charsets).
	//
	// MessageBody is a required field
	return nil
}
