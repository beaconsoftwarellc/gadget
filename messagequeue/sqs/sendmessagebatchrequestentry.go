package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

type smbreWrapper struct {
	*types.SendMessageBatchRequestEntry
}

func (smbi *smbreWrapper) SetQueueUrl(value string) {
	// log an error
	log.Warnf("SetQueueUrl called on unsupported struct SendMessageBatchRequestEntry")
}

func (smbi *smbreWrapper) SetMessageBody(value string) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.MessageBody = aws.String(value)
	}
}

func (smbi *smbreWrapper) SetDelaySeconds(value int32) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.DelaySeconds = value
	}
}

func (smbi *smbreWrapper) SetMessageAttributes(
	value map[string]types.MessageAttributeValue) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.MessageAttributes = value
	}
}

func (smbi *smbreWrapper) SetMessageSystemAttributes(
	value map[string]types.MessageSystemAttributeValue) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.MessageSystemAttributes = value
	}
}

func sendMessageBatchRequestEntryFromMessage(message *messagequeue.Message) (
	*types.SendMessageBatchRequestEntry, error) {
	wrapper := &smbreWrapper{
		SendMessageBatchRequestEntry: &types.SendMessageBatchRequestEntry{}}
	// ID is required and is used to match up request with response messages
	// it must be in (azAZ09_-), less than 80 characters, and unique
	// within the batch. We can just use generator.
	wrapper.SendMessageBatchRequestEntry.Id = aws.String(generator.String(32))
	return wrapper.SendMessageBatchRequestEntry,
		updateEnqueueFromMessage(wrapper, message)
}
