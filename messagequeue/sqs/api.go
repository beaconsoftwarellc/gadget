package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

//go:generate mockgen -package sqs -destination api_mock_test.gen.go . API

// API is an interface for mocking *sqs.SQS refer to the aws-sdk-go package for
// up to date documentation
type API interface {
	// SendMessage API operation for Amazon Simple Queue Service.
	//
	// Delivers a message to the specified queue.
	//
	// A message can include only XML, JSON, and unformatted text. The following
	// Unicode characters are allowed:
	//
	// #x9 | #xA | #xD | #x20 to #xD7FF | #xE000 to #xFFFD | #x10000 to #x10FFFF
	//
	// Any characters not included in this list will be rejected. For more information,
	// see the W3C specification for characters (http://www.w3.org/TR/REC-xml/#charsets).
	//
	// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
	// with awserr.Error's Code and Message methods to get detailed information about
	// the error.
	//
	// See the AWS API reference guide for Amazon Simple Queue Service's
	// API operation SendMessage for usage and error information.
	//
	// Returned Error Codes:
	//   * ErrCodeInvalidMessageContents "InvalidMessageContents"
	//   The message contains characters outside the allowed set.
	//
	//   * ErrCodeUnsupportedOperation "AWS.SimpleQueueService.UnsupportedOperation"
	//   Error code 400. Unsupported operation.
	//
	// See also, https://docs.aws.amazon.com/goto/WebAPI/sqs-2012-11-05/SendMessage
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (
		*sqs.SendMessageOutput, error)
	// SendMessageBatch API operation for Amazon Simple Queue Service.
	//
	// Delivers up to ten messages to the specified queue. This is a batch version
	// of SendMessage. For a FIFO queue, multiple messages within a single batch
	// are enqueued in the order they are sent.
	//
	// The result of sending each message is reported individually in the response.
	// Because the batch request can result in a combination of successful and unsuccessful
	// actions, you should check for batch errors even when the call returns an
	// HTTP status code of 200.
	//
	// The maximum allowed individual message size and the maximum total payload
	// size (the sum of the individual lengths of all of the batched messages) are
	// both 256 KB (262,144 bytes).
	//
	// A message can include only XML, JSON, and unformatted text. The following
	// Unicode characters are allowed:
	//
	// #x9 | #xA | #xD | #x20 to #xD7FF | #xE000 to #xFFFD | #x10000 to #x10FFFF
	//
	// Any characters not included in this list will be rejected. For more information,
	// see the W3C specification for characters (http://www.w3.org/TR/REC-xml/#charsets).
	//
	// If you don't specify the DelaySeconds parameter for an entry, Amazon SQS
	// uses the default value for the queue.
	//
	// Some actions take lists of parameters. These lists are specified using the
	// param.n notation. Values of n are integers starting from 1. For example,
	// a parameter list with two elements looks like this:
	//
	// &AttributeName.1=first
	//
	// &AttributeName.2=second
	//
	// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
	// with awserr.Error's Code and Message methods to get detailed information about
	// the error.
	//
	// See the AWS API reference guide for Amazon Simple Queue Service's
	// API operation SendMessageBatch for usage and error information.
	//
	// Returned Error Codes:
	//   * ErrCodeTooManyEntriesInBatchRequest "AWS.SimpleQueueService.TooManyEntriesInBatchRequest"
	//   The batch request contains more entries than permissible.
	//
	//   * ErrCodeEmptyBatchRequest "AWS.SimpleQueueService.EmptyBatchRequest"
	//   The batch request doesn't contain any entries.
	//
	//   * ErrCodeBatchEntryIdsNotDistinct "AWS.SimpleQueueService.BatchEntryIdsNotDistinct"
	//   Two or more batch entries in the request have the same Id.
	//
	//   * ErrCodeBatchRequestTooLong "AWS.SimpleQueueService.BatchRequestTooLong"
	//   The length of all the messages put together is more than the limit.
	//
	//   * ErrCodeInvalidBatchEntryId "AWS.SimpleQueueService.InvalidBatchEntryId"
	//   The Id of a batch entry in a batch request doesn't abide by the specification.
	//
	//   * ErrCodeUnsupportedOperation "AWS.SimpleQueueService.UnsupportedOperation"
	//   Error code 400. Unsupported operation.
	//
	// See also, https://docs.aws.amazon.com/goto/WebAPI/sqs-2012-11-05/SendMessageBatch
	SendMessageBatch(context.Context, *sqs.SendMessageBatchInput,
		...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
	// ReceiveMessage API operation for Amazon Simple Queue Service.
	//
	// Retrieves one or more messages (up to 10), from the specified queue. Using
	// the WaitTimeSeconds parameter enables long-poll support. For more information,
	// see Amazon SQS Long Polling (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-long-polling.html)
	// in the Amazon SQS Developer Guide.
	//
	// Short poll is the default behavior where a weighted random set of machines
	// is sampled on a ReceiveMessage call. Thus, only the messages on the sampled
	// machines are returned. If the number of messages in the queue is small (fewer
	// than 1,000), you most likely get fewer messages than you requested per ReceiveMessage
	// call. If the number of messages in the queue is extremely small, you might
	// not receive any messages in a particular ReceiveMessage response. If this
	// happens, repeat the request.
	//
	// For each message returned, the response includes the following:
	//
	//    * The message body.
	//
	//    * An MD5 digest of the message body. For information about MD5, see RFC1321
	//    (https://www.ietf.org/rfc/rfc1321.txt).
	//
	//    * The MessageId you received when you sent the message to the queue.
	//
	//    * The receipt handle.
	//
	//    * The message attributes.
	//
	//    * An MD5 digest of the message attributes.
	//
	// The receipt handle is the identifier you must provide when deleting the message.
	// For more information, see Queue and Message Identifiers (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-queue-message-identifiers.html)
	// in the Amazon SQS Developer Guide.
	//
	// You can provide the VisibilityTimeout parameter in your request. The parameter
	// is applied to the messages that Amazon SQS returns in the response. If you
	// don't include the parameter, the overall visibility timeout for the queue
	// is used for the returned messages. For more information, see Visibility Timeout
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)
	// in the Amazon SQS Developer Guide.
	//
	// A message that isn't deleted or a message whose visibility isn't extended
	// before the visibility timeout expires counts as a failed receive. Depending
	// on the configuration of the queue, the message might be sent to the dead-letter
	// queue.
	//
	// In the future, new attributes might be added. If you write code that calls
	// this action, we recommend that you structure your code so that it can handle
	// new attributes gracefully.
	//
	// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
	// with awserr.Error's Code and Message methods to get detailed information about
	// the error.
	//
	// See the AWS API reference guide for Amazon Simple Queue Service's
	// API operation ReceiveMessage for usage and error information.
	//
	// Returned Error Codes:
	//   * ErrCodeOverLimit "OverLimit"
	//   The specified action violates a limit. For example, ReceiveMessage returns
	//   this error if the maximum number of inflight messages is reached and AddPermission
	//   returns this error if the maximum number of permissions for the queue is
	//   reached.
	//
	// See also, https://docs.aws.amazon.com/goto/WebAPI/sqs-2012-11-05/ReceiveMessage
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (
		*sqs.ReceiveMessageOutput, error)
	// DeleteMessage API operation for Amazon Simple Queue Service.
	//
	// Deletes the specified message from the specified queue. To select the message
	// to delete, use the ReceiptHandle of the message (not the MessageId which
	// you receive when you send the message). Amazon SQS can delete a message from
	// a queue even if a visibility timeout setting causes the message to be locked
	// by another consumer. Amazon SQS automatically deletes messages left in a
	// queue longer than the retention period configured for the queue.
	//
	// The ReceiptHandle is associated with a specific instance of receiving a message.
	// If you receive a message more than once, the ReceiptHandle is different each
	// time you receive a message. When you use the DeleteMessage action, you must
	// provide the most recently received ReceiptHandle for the message (otherwise,
	// the request succeeds, but the message might not be deleted).
	//
	// For standard queues, it is possible to receive a message even after you delete
	// it. This might happen on rare occasions if one of the servers which stores
	// a copy of the message is unavailable when you send the request to delete
	// the message. The copy remains on the server and might be returned to you
	// during a subsequent receive request. You should ensure that your application
	// is idempotent, so that receiving a message more than once does not cause
	// issues.
	//
	// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
	// with awserr.Error's Code and Message methods to get detailed information about
	// the error.
	//
	// See the AWS API reference guide for Amazon Simple Queue Service's
	// API operation DeleteMessage for usage and error information.
	//
	// Returned Error Codes:
	//   * ErrCodeInvalidIdFormat "InvalidIdFormat"
	//   The specified receipt handle isn't valid for the current version.
	//
	//   * ErrCodeReceiptHandleIsInvalid "ReceiptHandleIsInvalid"
	//   The specified receipt handle isn't valid.
	//
	// See also, https://docs.aws.amazon.com/goto/WebAPI/sqs-2012-11-05/DeleteMessage
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}
