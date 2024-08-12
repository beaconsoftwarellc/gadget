package sqs

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func NewMatcher(assert *assert.Assertions,
	messages ...*messagequeue.Message) *MessageMatches {
	return &MessageMatches{assert: assert,
		messages: messages, smbo: &sqs.SendMessageBatchOutput{
			Successful: make([]types.SendMessageBatchResultEntry, len(messages)),
		}}
}

type MessageMatches struct {
	smbo     *sqs.SendMessageBatchOutput
	assert   *assert.Assertions
	messages []*messagequeue.Message
}

func getAttribute(m map[string]string, name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return v
	}
	return ""
}

func getAttributeValue(m map[string]types.MessageAttributeValue, name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return aws.ToString(v.StringValue)
	}
	return ""
}

func getSystemAttributeValue(
	m map[string]types.MessageSystemAttributeValue,
	name string) string {
	if nil == m {
		return ""
	}
	v, ok := m[name]
	if ok {
		return aws.ToString(v.StringValue)
	}
	return ""
}

func (mm *MessageMatches) Matches(arg interface{}) bool {
	switch o := arg.(type) {
	case *sqs.SendMessageBatchInput:
		for i, entry := range o.Entries {
			if i < len(mm.smbo.Successful) {
				mm.smbo.Successful[i].Id = entry.Id
			} else {
				mm.smbo.Failed[i-len(mm.smbo.Successful)].Id = entry.Id
			}
			if !mm.matches(mm.messages[i], entry) {
				return false
			}
		}
		return true
	default:
		return mm.matches(mm.messages[0], arg)
	}
}

func (mm *MessageMatches) matches(message *messagequeue.Message, arg interface{}) bool {
	var (
		delaySeconds int32
		external     string
		service      string
		method       string
		body         string
		trace        string
	)
	switch o := arg.(type) {
	case *sqs.SendMessageInput:
		delaySeconds = o.DelaySeconds
		service = getAttributeValue(o.MessageAttributes, serviceAttributeName)
		method = getAttributeValue(o.MessageAttributes, methodAttributeName)
		trace = getSystemAttributeValue(o.MessageSystemAttributes, awsTraceHeaderName)
		body = aws.ToString(o.MessageBody)
	case types.SendMessageBatchRequestEntry:
		entry := o
		delaySeconds = entry.DelaySeconds
		service = getAttributeValue(entry.MessageAttributes, serviceAttributeName)
		method = getAttributeValue(entry.MessageAttributes, methodAttributeName)
		trace = getSystemAttributeValue(entry.MessageSystemAttributes, awsTraceHeaderName)
		body = aws.ToString(entry.MessageBody)
	case *types.Message:
		service = getAttribute(o.Attributes, serviceAttributeName)
		method = getAttribute(o.Attributes, methodAttributeName)
		body = aws.ToString(o.Body)
	case *sqs.DeleteMessageInput:
		external = aws.ToString(o.ReceiptHandle)
	}
	return mm.assert.Equal(int32(message.Delay.Seconds()), delaySeconds) &&
		mm.assert.Equal(message.Service, service) &&
		mm.assert.Equal(message.Method, method) &&
		mm.assert.Equal(message.Body, body) &&
		mm.assert.Equal(message.Trace, trace) &&
		mm.assert.Equal(message.External, external)
}

func (mm *MessageMatches) String() string {
	return fmt.Sprintf("MessageMatches(%s)", "*mm.message")
}

func initialize(t *testing.T) (context.Context, *assert.Assertions, *MockAPI, *sdk) {
	assert := assert.New(t)
	context := context.Background()
	ctrl := gomock.NewController(t)
	apiMock := NewMockAPI(ctrl)
	qlocator := &url.URL{}
	obj := New("", qlocator)
	sdk := obj.(*sdk)
	sdk.api = apiMock
	return context, assert, apiMock, sdk
}

func Test_SQS_EnqueueBatch(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{}

	actual, actualError := sdk.EnqueueBatch(ctx, []*messagequeue.Message{message})
	assert.NoError(actualError)
	assert.Equal(1, len(actual))
	assert.Equal(message, actual[0].Message)
	assert.False(actual[0].Success)
	assert.Equal("invalid attribute value: minimum character count is 1", actual[0].Error)

	message.Service = generator.String(5)
	message.Method = generator.String(10)
	message.Body = generator.String(15)

	matcher := NewMatcher(assert, message)
	expectedID := generator.String(32)
	matcher.smbo = &sqs.SendMessageBatchOutput{
		Successful: []types.SendMessageBatchResultEntry{
			{MessageId: &expectedID},
		}}
	apiMock.EXPECT().SendMessageBatch(ctx, matcher, gomock.Any()).
		Return(matcher.smbo, nil)
	actual, actualError = sdk.EnqueueBatch(ctx, []*messagequeue.Message{message})
	assert.NoError(actualError)
	assert.Equal(1, len(actual))
	assert.Equal(message, actual[0].Message)
	assert.Equal(expectedID, message.ID)
	assert.True(actual[0].Success)
}

func Test_SQS_EnqueueBatch_Failure(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{}
	message.Service = generator.String(5)
	message.Method = generator.String(10)
	message.Body = generator.String(15)

	matcher := NewMatcher(assert, message)
	expectedFailMessage := generator.String(32)
	expectedFailCode := generator.String(2)
	expectedSenderFault := true
	matcher.smbo = &sqs.SendMessageBatchOutput{
		Successful: []types.SendMessageBatchResultEntry{},
		Failed: []types.BatchResultErrorEntry{
			{
				SenderFault: expectedSenderFault,
				Message:     &expectedFailMessage,
				Code:        &expectedFailCode,
			},
		},
	}
	apiMock.EXPECT().SendMessageBatch(ctx, matcher, gomock.Any()).
		Return(matcher.smbo, nil)
	actual, actualError := sdk.EnqueueBatch(ctx, []*messagequeue.Message{message})
	assert.NoError(actualError)
	assert.Equal(1, len(actual))
	assert.Equal(message, actual[0].Message)
	assert.False(actual[0].Success)
	assert.Equal(expectedSenderFault, actual[0].SenderFault)
	assert.Equal(fmt.Sprintf("%s: %s", expectedFailCode, expectedFailMessage),
		actual[0].Error)
}

func Test_SQS_EnqueueBatch_AtMax(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	messages := make([]*messagequeue.Message, 10)
	for i := range messages {
		messages[i] = &messagequeue.Message{
			Service: generator.String(32),
			Method:  generator.String(32),
			Body:    generator.String(32),
		}
	}
	matcher := NewMatcher(assert, messages...)
	apiMock.EXPECT().SendMessageBatch(ctx, matcher, gomock.Any()).
		Return(matcher.smbo, nil)
	actual, actualError := sdk.EnqueueBatch(ctx, messages)
	assert.NoError(actualError)
	assert.Equal(10, len(actual))
}

func Test_SQS_EnqueueBatch_OverMax(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	messages := make([]*messagequeue.Message, 12)
	for i := range messages {
		messages[i] = &messagequeue.Message{
			Service: generator.String(32),
			Method:  generator.String(32),
			Body:    generator.String(32),
		}
	}
	matcher := NewMatcher(assert, messages[0:10]...)
	apiMock.EXPECT().SendMessageBatch(ctx, matcher, gomock.Any()).
		Return(matcher.smbo, nil)

	matcher2 := NewMatcher(assert, messages[10:]...)
	apiMock.EXPECT().SendMessageBatch(ctx, matcher2, gomock.Any()).
		Return(matcher2.smbo, nil)

	actual, actualError := sdk.EnqueueBatch(ctx, messages)
	assert.NoError(actualError)
	assert.Equal(12, len(actual))
}

func Test_SQS_Dequeue(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	var (
		wait              time.Duration = 0
		visibilityTimeout time.Duration = 0
		count             int           = 0
	)
	locator, _ := url.Parse(fmt.Sprintf("http://%s.com",
		strings.ToLower(generator.String(20))))
	sdk.queueUrl = locator
	apiMock.EXPECT().ReceiveMessage(ctx, gomock.Any(), gomock.Any()).Return(
		&sqs.ReceiveMessageOutput{Messages: []types.Message{}}, nil,
	)
	actual, err := sdk.Dequeue(ctx, count, wait, visibilityTimeout)
	assert.NoError(err)
	assert.Equal(0, len(actual))

	expectedMessage := &messagequeue.Message{
		ID:       generator.String(5),
		External: generator.String(10),
		Service:  missing,
		Method:   missing,
	}
	sqsMessage := types.Message{}
	sqsMessage.MessageId = aws.String(expectedMessage.ID)
	sqsMessage.ReceiptHandle = aws.String(expectedMessage.External)
	apiMock.EXPECT().ReceiveMessage(ctx, gomock.Any(), gomock.Any()).Return(
		&sqs.ReceiveMessageOutput{Messages: []types.Message{sqsMessage}}, nil,
	)
	actual, err = sdk.Dequeue(ctx, count, wait, visibilityTimeout)
	assert.NoError(err)
	assert.Equal(1, len(actual))
	assert.Equal(expectedMessage.ID, actual[0].ID)
	assert.Equal(expectedMessage.External, actual[0].External)
	assert.Equal(expectedMessage.Service, actual[0].Service)
	assert.Equal(expectedMessage.Method, actual[0].Method)
}

func Test_SQS_Delete(t *testing.T) {
	ctx, assert, apiMock, sdk := initialize(t)
	message := &messagequeue.Message{
		External: generator.String(32),
	}
	expected := generator.String(32)
	apiMock.EXPECT().DeleteMessage(ctx, NewMatcher(assert, message), gomock.Any()).
		Return(nil, errors.New(expected))
	assert.EqualError(sdk.Delete(ctx, message), expected)
}
