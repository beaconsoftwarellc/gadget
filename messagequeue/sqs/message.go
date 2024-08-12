package sqs

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

const missing = "missing"

func convert(msg *types.Message, deadline time.Time) *messagequeue.Message {
	mqMessage := &messagequeue.Message{
		ID:       aws.ToString(msg.MessageId),
		External: aws.ToString(msg.ReceiptHandle),
		Body:     aws.ToString(msg.Body),
		Service:  missing,
		Method:   missing,
		Deadline: deadline,
	}
	service, ok := msg.MessageAttributes[serviceAttributeName]
	if ok {
		mqMessage.Service = aws.ToString(service.StringValue)
	}
	method, ok := msg.MessageAttributes[methodAttributeName]
	if ok {
		mqMessage.Method = aws.ToString(method.StringValue)
	}
	return mqMessage
}
