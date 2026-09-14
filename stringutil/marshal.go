package stringutil

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/beaconsoftwarellc/gadget/v2/log"
	"github.com/beaconsoftwarellc/gadget/v2/messagequeue"
	"google.golang.org/protobuf/proto"
)

var (
	jsonNull       = "null"
	base64encoding = base64.StdEncoding.WithPadding(base64.StdPadding)
	payloadKey     = "Payload"
)

func base64decode(s string) ([]byte, bool) {
	// the error type doesn't matter
	// the string failed to decode, which is all we care about
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Debugf("[STR.MAR.25] failed to decode string: '%s', error: %s", s, err)
		return []byte(s), false
	}
	return decoded, true
}

// DebugDecode a message from a string using anonymous types. This method will decode the payload
// as well (if necessary) and set it to the appropriate value. The message will then be 'pretty' printed
// using JSON Marshal.
func DebugDecode(message *messagequeue.Message) (string, bool, error) {
	var (
		target map[string]interface{}
		body   []byte
		result []byte
		ok     bool
		err    error
	)
	body, ok = base64decode(message.Body)
	if !ok {
		return message.Body, ok, err
	}
	target = make(map[string]interface{})
	err = json.Unmarshal(body, &target)
	if err != nil {
		return "", false, err
	}
	if target[payloadKey] == nil {
		target[payloadKey] = "nil"
	} else {
		payload, _ := base64decode(target[payloadKey].(string))
		target[payloadKey] = string(payload)
	}
	result, err = json.MarshalIndent(target, "", "\t")
	if err != nil {
		log.ExitOnError(err)
	}
	return string(result), ok, nil
}

// EncodeMessage as a string
func EncodeMessage(message proto.Message) (string, error) {
	var (
		messageString string
		messageBytes  []byte
		err           error
	)
	messageBytes, err = json.Marshal(message)
	if nil != err {
		return messageString, err
	}
	if isJsonNull(messageBytes) {
		return messageString, errors.New("message was <nil>")
	}
	messageString = base64encoding.EncodeToString(messageBytes)
	return messageString, nil
}

// DecodeMessage from a string
func DecodeMessage(messageString string, target proto.Message) error {
	var (
		messageBytes []byte
		err          error
	)
	messageBytes, err = base64encoding.DecodeString(messageString)
	if nil != err {
		return err
	}
	if isJsonNull(messageBytes) {
		return errors.New("messageString is encoded <nil>")
	}
	return json.Unmarshal(messageBytes, target)
}

func isJsonNull(b []byte) bool {
	return len(b) == len(jsonNull) && string(b) == jsonNull
}
