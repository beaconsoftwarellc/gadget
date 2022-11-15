package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const missing = "missing"

func convert(msg *sqs.Message) *messagequeue.Message {
	mqMessage := &messagequeue.Message{
		ID:       *msg.MessageId,
		External: *msg.ReceiptHandle,
		Body:     *msg.Body,
		Service:  missing,
		Method:   missing,
	}
	service, ok := msg.MessageAttributes[serviceAttributeName]
	if ok {
		mqMessage.Service = *service.StringValue
	}
	method, ok := msg.MessageAttributes[methodAttributeName]
	if ok {
		mqMessage.Method = *method.StringValue
	}
	return mqMessage
}
