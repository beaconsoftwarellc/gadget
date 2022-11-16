package sqs

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func NewMatcher(assert *assert.Assertions, message *messagequeue.Message) gomock.Matcher {
	return &MessageMatches{assert: assert, message: message}
}

type MessageMatches struct {
	assert  *assert.Assertions
	message *messagequeue.Message
}

func getAttribute(m map[string]*string, name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return aws.StringValue(v)
	}
	return ""
}

func getAttributeValue(m map[string]*sqs.MessageAttributeValue, name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return aws.StringValue(v.StringValue)
	}
	return ""
}

func getSystemAttributeValue(m map[string]*sqs.MessageSystemAttributeValue, name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return aws.StringValue(v.StringValue)
	}
	return ""
}

func (mm *MessageMatches) Matches(arg interface{}) bool {
	var (
		delaySeconds int64
		external     string
		service      string
		method       string
		body         string
		trace        string
	)
	switch o := arg.(type) {
	case *sqs.SendMessageInput:
		delaySeconds = aws.Int64Value(o.DelaySeconds)
		service = getAttributeValue(o.MessageAttributes, serviceAttributeName)
		method = getAttributeValue(o.MessageAttributes, methodAttributeName)
		trace = getSystemAttributeValue(o.MessageSystemAttributes, awsTraceHeaderName)
		body = aws.StringValue(o.MessageBody)
	case *sqs.SendMessageBatchInput:
		if len(o.Entries) > 0 {
			entry := o.Entries[0]
			delaySeconds = aws.Int64Value(entry.DelaySeconds)
			service = getAttributeValue(entry.MessageAttributes, serviceAttributeName)
			method = getAttributeValue(entry.MessageAttributes, methodAttributeName)
			trace = getSystemAttributeValue(entry.MessageSystemAttributes, awsTraceHeaderName)
			body = aws.StringValue(entry.MessageBody)
		} else {
			return nil == mm.message
		}
	case *sqs.Message:
		service = getAttribute(o.Attributes, serviceAttributeName)
		method = getAttribute(o.Attributes, methodAttributeName)
		body = aws.StringValue(o.Body)
	case *sqs.DeleteMessageInput:
		external = aws.StringValue(o.ReceiptHandle)
	}
	return mm.assert.Equal(int64(mm.message.Delay.Seconds()), delaySeconds) &&
		mm.assert.Equal(mm.message.Service, service) &&
		mm.assert.Equal(mm.message.Method, method) &&
		mm.assert.Equal(mm.message.Body, body) &&
		mm.assert.Equal(mm.message.Trace, trace) &&
		mm.assert.Equal(mm.message.External, external)
}

func (mm *MessageMatches) String() string {
	return fmt.Sprintf("MessageMatches(%s)", "*mm.message")
}

func initialize(t *testing.T) (*assert.Assertions, *MockAPI, *sdk) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	apiMock := NewMockAPI(ctrl)
	qlocator := &url.URL{}
	obj := New(qlocator)
	sdk := obj.(*sdk)
	sdk.api = apiMock
	return assert, apiMock, sdk
}

func Test_SQS_Enqueue(t *testing.T) {
	assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{}

	actual := sdk.Enqueue(message)
	assert.EqualError(actual, "invalid attribute value: minimum character count is 1")

	message.Service = generator.String(5)
	message.Method = generator.String(10)
	message.Body = generator.String(15)

	expectedID := generator.String(32)
	apiMock.EXPECT().SendMessage(NewMatcher(assert, message)).
		Return(&sqs.SendMessageOutput{MessageId: &expectedID}, nil)
	actual = sdk.Enqueue(message)
	assert.NoError(actual)
	assert.Equal(message.ID, expectedID)
}

func Test_SQS_EnqueueBatch(t *testing.T) {
	assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{}

	actual := sdk.EnqueueBatch([]*messagequeue.Message{message})
	assert.EqualError(actual, "all messages were invalid")

	message.Service = generator.String(5)
	message.Method = generator.String(10)
	message.Body = generator.String(15)

	apiMock.EXPECT().SendMessageBatch(NewMatcher(assert, message)).
		Return(&sqs.SendMessageBatchOutput{}, nil)
	actual = sdk.EnqueueBatch([]*messagequeue.Message{message})
	assert.NoError(actual)
}

func Test_SQS_EnqueueBatch_Failure(t *testing.T) {
	assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{}

	actual := sdk.EnqueueBatch([]*messagequeue.Message{message})
	assert.EqualError(actual, "all messages were invalid")

	message.Service = generator.String(5)
	message.Method = generator.String(10)
	message.Body = generator.String(15)

	failure := &sqs.BatchResultErrorEntry{}
	failure.SetCode(generator.String(20))
	failure.SetSenderFault(true)
	failure.SetCode(generator.String(25))
	failure.SetCode(generator.String(25))
	apiMock.EXPECT().SendMessageBatch(NewMatcher(assert, message)).
		Return(&sqs.SendMessageBatchOutput{Failed: []*sqs.BatchResultErrorEntry{failure}}, nil)
	actual = sdk.EnqueueBatch([]*messagequeue.Message{message})
	assert.NoError(actual)
}

func Test_SQS_Dequeue(t *testing.T) {
	assert, apiMock, sdk := initialize(t)
	var (
		wait     time.Duration = 0
		count    int           = 0
		expected               = &sqs.ReceiveMessageInput{}
	)
	locator, _ := url.Parse(fmt.Sprintf("http://%s.com", strings.ToLower(generator.String(20))))
	sdk.queueUrl = locator
	expected.SetQueueUrl(locator.String())
	expected.SetMaxNumberOfMessages(1)
	expected.SetWaitTimeSeconds(0)
	apiMock.EXPECT().ReceiveMessage(expected).Return(
		&sqs.ReceiveMessageOutput{Messages: []*sqs.Message{}}, nil,
	)
	actual, err := sdk.Dequeue(count, wait)
	assert.NoError(err)
	assert.Equal(0, len(actual))

	expectedMessage := &messagequeue.Message{
		ID:       generator.String(5),
		External: generator.String(10),
		Service:  missing,
		Method:   missing,
	}
	sqsMessage := &sqs.Message{}
	sqsMessage.SetMessageId(expectedMessage.ID)
	sqsMessage.SetReceiptHandle(expectedMessage.External)
	apiMock.EXPECT().ReceiveMessage(expected).Return(
		&sqs.ReceiveMessageOutput{Messages: []*sqs.Message{sqsMessage}}, nil,
	)
	actual, err = sdk.Dequeue(count, wait)
	assert.NoError(err)
	assert.Equal(1, len(actual))
	assert.Equal(expectedMessage, actual[0])
}

func Test_SQS_Delete(t *testing.T) {
	assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{
		External: generator.String(32),
	}
	expected := generator.String(32)
	apiMock.EXPECT().DeleteMessage(NewMatcher(assert, message)).
		Return(nil, errors.New(expected))
	assert.EqualError(sdk.Delete(message), expected)
}
