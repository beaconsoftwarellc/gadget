package stringutil

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	jsonNull       = "null"
	base64encoding = base64.StdEncoding.WithPadding(base64.StdPadding)
)

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
