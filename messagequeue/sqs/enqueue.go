package sqs

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const (
	stringDataType         = "String"
	maxDelay               = 15 * time.Minute
	mappingErrorMessage    = "mapping must not be nil"
	messageNilErrorMessage = "message cannot be nil"
)

type enqueue interface {
	SetQueueUrl(string)
	SetMessageBody(string)
	// The length of time, in seconds, for which to delay a specific message. Valid
	// values: 0 to 900. Maximum: 15 minutes. Messages with a positive DelaySeconds
	// value become available for processing after the delay period is finished.
	// If you don't specify a value, the default value for the queue applies.
	//
	// When you set FifoQueue, you can't set DelaySeconds per message. You can set
	// this parameter only on a queue level.
	SetDelaySeconds(int64)
	SetMessageAttributes(map[string]*sqs.MessageAttributeValue)
	SetMessageSystemAttributes(map[string]*sqs.MessageSystemAttributeValue)
}

func updateEnqueueFromMessage(request enqueue, message *messagequeue.Message) error {
	if nil == message {
		return errors.New("message cannot be nil")
	}
	var (
		err  error
		mma  = map[string]*sqs.MessageAttributeValue{}
		mmsa = map[string]*sqs.MessageSystemAttributeValue{}
	)
	if message.Delay > maxDelay {
		message.Delay = maxDelay
	}
	request.SetDelaySeconds(int64(message.Delay.Seconds()))

	if err = setAttribute(mma, serviceAttributeName, message.Service); err != nil {
		return err
	}

	if err = setAttribute(mma, methodAttributeName, message.Method); err != nil {
		return err
	}

	request.SetMessageAttributes(mma)

	if err = BodyIsValid(message.Body); err != nil {
		return err
	}
	request.SetMessageBody(message.Body)

	if !stringutil.IsWhiteSpace(message.Trace) {
		err = setXRayTrace(mmsa, message.Trace)
	}
	request.SetMessageSystemAttributes(mmsa)
	return err
}

func setAttribute(mapping map[string]*sqs.MessageAttributeValue, name, value string) error {
	if nil == mapping {
		return errors.New(mappingErrorMessage)
	}
	var err error
	if err = NameIsValid(name); nil != err {
		return err
	}
	if err = BodyIsValid(value); nil != err {
		return errors.New("invalid attribute value: %s", err.Error())
	}
	mav := &sqs.MessageAttributeValue{}
	mav.SetDataType(stringDataType)
	mav.SetStringValue(value)
	mapping[name] = mav
	return nil
}

func setXRayTrace(mapping map[string]*sqs.MessageSystemAttributeValue,
	value string) error {
	if nil == mapping {
		return errors.New(mappingErrorMessage)
	}
	// the only supported message system attribute is AWSTraceHeader.
	// Its type must be String and its value must be a correctly formatted X-Ray
	// trace header string.
	var err error
	if err = BodyIsValid(value); nil != err {
		return err
	}
	msav := &sqs.MessageSystemAttributeValue{}
	msav.SetDataType(stringDataType)
	msav.SetStringValue(value)
	mapping[awsTraceHeaderName] = msav
	return nil
}
