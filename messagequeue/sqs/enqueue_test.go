package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_updateEnqueueFromMessage(t *testing.T) {
	var (
		assert  = assert1.New(t)
		actual  error
		message *messagequeue.Message
		enqueue = &smiWrapper{SendMessageInput: &sqs.SendMessageInput{}}
	)
	actual = updateEnqueueFromMessage(enqueue, message)
	assert.EqualError(actual, messageNilErrorMessage)

	message = &messagequeue.Message{}
	actual = updateEnqueueFromMessage(enqueue, message)
	assert.EqualError(actual, "invalid attribute value: minimum character count is 1")

	message.Service = generator.String(32)
	actual = updateEnqueueFromMessage(enqueue, message)
	assert.EqualError(actual, "invalid attribute value: minimum character count is 1")

	message.Method = generator.String(32)
	actual = updateEnqueueFromMessage(enqueue, message)
	assert.EqualError(actual, "minimum character count is 1")

	message.Trace = generator.String(32)
	message.Body = generator.String(32)
	actual = updateEnqueueFromMessage(enqueue, message)
	assert.NoError(actual)

	assert.Equal(message.Body, aws.ToString(enqueue.MessageBody))
	assert.Equal(int32(message.Delay.Seconds()), enqueue.DelaySeconds)
	assert.Equal(message.Service,
		*enqueue.MessageAttributes[serviceAttributeName].StringValue)
	assert.Equal(message.Method,
		*enqueue.MessageAttributes[methodAttributeName].StringValue)

}

func Test_setAttribute(t *testing.T) {
	var (
		assert   = assert1.New(t)
		mapping  map[string]types.MessageAttributeValue
		name     string
		value    string
		expected string
		actual   error
	)
	expected = mappingErrorMessage
	actual = setAttribute(mapping, name, value)
	assert.EqualError(actual, expected)

	mapping = make(map[string]types.MessageAttributeValue)
	expected = "name character count out of bounds [1, 256] (0)"
	actual = setAttribute(mapping, name, value)
	assert.EqualError(actual, expected)

	name = generator.String(32)
	actual = setAttribute(mapping, name, value)
	assert.EqualError(actual, "invalid attribute value: minimum character count is 1")

	value = generator.String(32)
	actual = setAttribute(mapping, name, value)
	assert.NoError(actual)
	assert.Equal(value, *mapping[name].StringValue)
	assert.Equal(stringDataType, *mapping[name].DataType)
}

func Test_setXRayTrace(t *testing.T) {
	var (
		assert   = assert1.New(t)
		mapping  map[string]types.MessageSystemAttributeValue
		value    string
		expected string
		actual   error
	)
	expected = mappingErrorMessage
	actual = setXRayTrace(mapping, value)
	assert.EqualError(actual, expected)

	mapping = make(map[string]types.MessageSystemAttributeValue)
	expected = bodyMinimumError
	actual = setXRayTrace(mapping, value)
	assert.EqualError(actual, expected)

	value = generator.String(32)
	actual = setXRayTrace(mapping, value)
	assert.NoError(actual)

	assert.Equal(value, *mapping[awsTraceHeaderName].StringValue)
	assert.Equal(stringDataType, *mapping[awsTraceHeaderName].DataType)
}
