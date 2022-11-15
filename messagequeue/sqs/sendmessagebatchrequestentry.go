package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
)

type smbreWrapper struct {
	*sqs.SendMessageBatchRequestEntry
}

func (smbi *smbreWrapper) SetQueueUrl(value string) {
	// log an error
	log.Warnf("SetQueueUrl called on unsupported struct SendMessageBatchRequestEntry")
}

func (smbi *smbreWrapper) SetMessageBody(value string) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.SetMessageBody(value)
	}
}

func (smbi *smbreWrapper) SetDelaySeconds(value int64) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.SetDelaySeconds(value)
	}
}

func (smbi *smbreWrapper) SetMessageAttributes(
	value map[string]*sqs.MessageAttributeValue) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.SetMessageAttributes(value)
	}
}

func (smbi *smbreWrapper) SetMessageSystemAttributes(
	value map[string]*sqs.MessageSystemAttributeValue) {
	if nil != smbi && nil != smbi.SendMessageBatchRequestEntry {
		smbi.SendMessageBatchRequestEntry.SetMessageSystemAttributes(value)
	}
}

func sendMessageBatchRequestEntryFromMessage(message *messagequeue.Message) (
	*sqs.SendMessageBatchRequestEntry, error) {
	wrapper := &smbreWrapper{
		SendMessageBatchRequestEntry: &sqs.SendMessageBatchRequestEntry{}}
	// ID is required and is used to match up request with response messages
	// it must be in (azAZ09_-), less than 80 characters, and unique
	// within the batch. We can just use generator.
	wrapper.SendMessageBatchRequestEntry.SetId(generator.String(32))
	return wrapper.SendMessageBatchRequestEntry,
		updateEnqueueFromMessage(wrapper, message)
}
