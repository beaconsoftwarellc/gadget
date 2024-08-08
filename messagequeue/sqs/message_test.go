package sqs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	assert1 "github.com/stretchr/testify/assert"
)

func Test_convert(t *testing.T) {
	assert := assert1.New(t)
	var (
		id       string
		external string
		body     string
		service  string = missing
		method   string = missing
		deadline        = time.Now().Add(time.Second)
	)
	message := &types.Message{
		MessageAttributes: make(map[string]types.MessageAttributeValue),
	}
	assertAll := func(actual *messagequeue.Message) {
		assert.Equal(id, actual.ID)
		assert.Equal(body, actual.Body)
		assert.Equal(external, actual.External)
		assert.Equal(service, actual.Service)
		assert.Equal(method, actual.Method)
		assert.Equal(deadline, actual.Deadline)
	}
	actual := convert(message, deadline)
	assertAll(actual)

	id = generator.String(2)
	external = generator.String(3)
	body = generator.String(4)
	service = generator.String(5)
	method = generator.String(6)
	message.MessageId = &id
	message.ReceiptHandle = &external
	message.Body = &body
	message.MessageAttributes[serviceAttributeName] = types.MessageAttributeValue{
		StringValue: &service,
	}
	message.MessageAttributes[methodAttributeName] = types.MessageAttributeValue{
		StringValue: &method,
	}
	actual = convert(message, deadline)
	assertAll(actual)
}
