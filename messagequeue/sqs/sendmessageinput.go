package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const (
	minNameCharacters = 1
	maxNameCharacters = 256
	minBodyCharacters = 1
	maxBodyKilobytes  = 255
	prohibitedAWS     = "aws"
	prohibitedAmazon  = "amazon"
	period            = "."
	stringDataType    = "String"
)

func sendMessageInputFromMessage(message *messagequeue.Message) (*sqs.SendMessageInput, error) {
	var err error
	smi := &sqs.SendMessageInput{
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/using-messagededuplicationid-property.html
		MessageDeduplicationId: nil,
		// This parameter applies only to FIFO (first-in-first-out) queues.
		// Message Group ID guarantee's ORDERING within the GROUP ID
		// To  avoid processing duplicate messages in a system with multiple producers
		// and consumers where throughput and latency are more important than
		// ordering, the producer should generate a unique message group ID for
		// each message.
		// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/using-messagegroupid-property.html
		MessageGroupId:    nil,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{},
		// the only supported message system attribute is AWSTraceHeader.
		// Its type must be String and its value must be a correctly formatted X-Ray
		// trace header string.
		MessageSystemAttributes: map[string]*sqs.MessageSystemAttributeValue{},
	}

	if err = setAttribute(smi, serviceAttributeName, message.Service); err != nil {
		return nil, err
	}

	if err = setAttribute(smi, methodAttributeName, message.Method); err != nil {
		return nil, err
	}

	if err = BodyIsValid(message.Body); err != nil {
		return nil, err
	}
	smi.SetMessageBody(message.Body)

	if !stringutil.IsWhiteSpace(message.Trace) {
		err = setXRayTrace(smi, message.Trace)
	}
	return smi, err
}

func setAttribute(smi *sqs.SendMessageInput, name, value string) error {
	var err error
	if err = NameIsValid(name); nil != err {
		return err
	}
	if err = BodyIsValid(value); nil != err {
		return err
	}
	mav := &sqs.MessageAttributeValue{}
	mav.SetDataType(stringDataType)
	mav.SetStringValue(value)
	smi.MessageAttributes[name] = mav
	return nil
}

func setXRayTrace(smi *sqs.SendMessageInput, value string) error {
	var err error
	if err = BodyIsValid(value); nil != err {
		return err
	}
	msav := &sqs.MessageSystemAttributeValue{}
	msav.SetDataType(stringDataType)
	msav.SetStringValue(value)
	smi.MessageSystemAttributes[awsTraceHeaderName] = msav
}
