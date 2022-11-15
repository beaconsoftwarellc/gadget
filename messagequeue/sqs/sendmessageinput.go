package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

type smiWrapper struct {
	*sqs.SendMessageInput
}

func (smi *smiWrapper) SetQueueUrl(value string) {
	smi.SendMessageInput.SetQueueUrl(value)
}

func (smi *smiWrapper) SetMessageBody(value string) {
	smi.SendMessageInput.SetMessageBody(value)
}

func (smi *smiWrapper) SetDelaySeconds(value int64) {
	smi.SendMessageInput.SetDelaySeconds(value)
}

func (smi *smiWrapper) SetMessageAttributes(value map[string]*sqs.MessageAttributeValue) {
	smi.SendMessageInput.SetMessageAttributes(value)
}

func (smi *smiWrapper) SetMessageSystemAttributes(value map[string]*sqs.MessageSystemAttributeValue) {
	smi.SendMessageInput.SetMessageSystemAttributes(value)
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
	return wrapper.SendMessageInput, updateRequestFromMessage(wrapper, message)
}
