package sqs

import (
	"testing"
	"time"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_sendMessageInputFromMessage(t *testing.T) {
	assert := assert1.New(t)
	message := &messagequeue.Message{
		ID:       generator.String(32),
		External: generator.String(32),
		Trace:    generator.String(32),
		Delay:    time.Second,
		Service:  generator.String(32),
		Method:   generator.String(32),
		Body:     generator.String(32),
	}
	actual, err := sendMessageInputFromMessage(message)
	assert.NoError(err)
	assert.Equal(message.Body, *actual.MessageBody)
	assert.Equal(message.Service,
		*actual.MessageAttributes[serviceAttributeName].StringValue)
	assert.Equal(message.Method,
		*actual.MessageAttributes[methodAttributeName].StringValue)
	assert.Equal(message.Trace,
		*actual.MessageSystemAttributes[awsTraceHeaderName].StringValue)
	assert.Equal(int32(message.Delay.Seconds()), actual.DelaySeconds)
}
