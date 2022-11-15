package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const missing = "missing"

func convert(msg *sqs.Message) *messagequeue.Message {
	mqMessage := &messagequeue.Message{
		ID:       aws.StringValue(msg.MessageId),
		External: aws.StringValue(msg.ReceiptHandle),
		Body:     aws.StringValue(msg.Body),
		Service:  missing,
		Method:   missing,
	}
	service, ok := msg.MessageAttributes[serviceAttributeName]
	if ok {
		mqMessage.Service = aws.StringValue(service.StringValue)
	}
	method, ok := msg.MessageAttributes[methodAttributeName]
	if ok {
		mqMessage.Method = aws.StringValue(method.StringValue)
	}
	return mqMessage
}
