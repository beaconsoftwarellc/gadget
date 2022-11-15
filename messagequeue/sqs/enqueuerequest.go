package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"github.com/beaconsoftwarellc/gadget/v2/stringutil"
)

const stringDataType = "String"

type enqueueRequest interface {
	SetQueueUrl(string)
	SetMessageBody(string)
	SetDelaySeconds(int64)
	SetMessageAttributes(map[string]*sqs.MessageAttributeValue)
	SetMessageSystemAttributes(map[string]*sqs.MessageSystemAttributeValue)
}

func updateRequestFromMessage(request enqueueRequest, message *messagequeue.Message) error {
	var (
		err  error
		mma  = map[string]*sqs.MessageAttributeValue{}
		mmsa = map[string]*sqs.MessageSystemAttributeValue{}
	)

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
	mapping[name] = mav
	return nil
}

func setXRayTrace(mapping map[string]*sqs.MessageSystemAttributeValue,
	value string) error {
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
