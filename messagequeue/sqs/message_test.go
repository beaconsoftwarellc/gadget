package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
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
	)
	message := &sqs.Message{
		MessageAttributes: make(map[string]*sqs.MessageAttributeValue),
	}
	assertAll := func(actual *messagequeue.Message) {
		assert.Equal(id, actual.ID)
		assert.Equal(body, actual.Body)
		assert.Equal(external, actual.External)
		assert.Equal(service, actual.Service)
		assert.Equal(method, actual.Method)
	}
	actual := convert(message)
	assertAll(actual)

	id = generator.String(2)
	external = generator.String(3)
	body = generator.String(4)
	service = generator.String(5)
	method = generator.String(6)
	message.MessageId = &id
	message.ReceiptHandle = &external
	message.Body = &body
	message.MessageAttributes[serviceAttributeName] = &sqs.MessageAttributeValue{
		StringValue: &service,
	}
	message.MessageAttributes[methodAttributeName] = &sqs.MessageAttributeValue{
		StringValue: &method,
	}
	actual = convert(message)
	assertAll(actual)
}
