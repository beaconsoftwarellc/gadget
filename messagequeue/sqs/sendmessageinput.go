package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

type smiWrapper struct {
	*sqs.SendMessageInput
}

func (smi *smiWrapper) SetQueueUrl(value string) {
	if nil != smi && nil != smi.SendMessageInput {
		smi.SendMessageInput.QueueUrl = aws.String(value)
	}
}

func (smi *smiWrapper) SetMessageBody(value string) {
	if nil != smi && nil != smi.SendMessageInput {
		smi.SendMessageInput.MessageBody = aws.String(value)
	}
}

func (smi *smiWrapper) SetDelaySeconds(value int32) {
	if nil != smi && nil != smi.SendMessageInput {
		smi.SendMessageInput.DelaySeconds = value
	}
}

func (smi *smiWrapper) SetMessageAttributes(value map[string]types.MessageAttributeValue) {
	if nil != smi && nil != smi.SendMessageInput {
		smi.SendMessageInput.MessageAttributes = value
	}
}

func (smi *smiWrapper) SetMessageSystemAttributes(value map[string]types.MessageSystemAttributeValue) {
	if nil != smi && nil != smi.SendMessageInput {
		smi.SendMessageInput.MessageSystemAttributes = value
	}
}

func sendMessageInputFromMessage(message *messagequeue.Message) (*sqs.SendMessageInput, error) {
	wrapper := &smiWrapper{SendMessageInput: &sqs.SendMessageInput{
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/using-messagededuplicationid-property.html
		MessageDeduplicationId: nil,
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/using-messagegroupid-property.html
		MessageGroupId: nil,
	}}
	return wrapper.SendMessageInput, updateEnqueueFromMessage(wrapper, message)
}
