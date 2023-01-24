package stringutil

import (
	"testing"

	"time"

	assert1 "github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type TestRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID       string                 `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID"`
	Name     string                 `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name"`
	Created  *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=Created,proto3" json:"Created"`
	Modified *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=Modified,proto3" json:"Modified"`
}

func (x *TestRecord) ProtoReflect() protoreflect.Message {
	var obj protoimpl.MessageInfo
	obj.Exporter = func(v interface{}, i int) interface{} {
		switch v := v.(*TestRecord); i {
		case 0:
			return &v.state
		case 1:
			return &v.sizeCache
		case 2:
			return &v.unknownFields
		default:
			return nil
		}
	}
	mi := &obj
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func TestEncodeMessage(t *testing.T) {
	var tests = []struct {
		name          string
		message       proto.Message
		expected      string
		expectedError string
	}{
		{
			name:          "nil does not explode",
			message:       nil,
			expected:      "",
			expectedError: "message was <nil>",
		},
		{
			name:    "trivial encode",
			message: &TestRecord{},
			expected: "eyJJRCI6IiIsIk5hbWUiOiIiLCJDcmVhdGVkIjpudWxsLCJNb2R" +
				"pZmllZCI6bnVsbH0=",
			expectedError: "",
		},
		{
			name: "non-trivial encode",
			message: &TestRecord{
				ID:       "12345",
				Name:     "non-trivial encode",
				Created:  timestamppb.New(time.Unix(0, 0)),
				Modified: timestamppb.New(time.Unix(1337, 0)),
			},
			expected: "eyJJRCI6IjEyMzQ1IiwiTmFtZSI6Im5vbi10cml2aWFsIGVuY29k" +
				"ZSIsIkNyZWF0ZWQiOnt9LCJNb2RpZmllZCI6eyJzZWNvbmRzIjoxMzM3fX0=",
			expectedError: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			actual, actualErr := EncodeMessage(tc.message)
			if IsWhiteSpace(tc.expectedError) {
				assert.Equal(tc.expected, actual)
			} else {
				assert.EqualError(actualErr, tc.expectedError)
			}
		})
	}
}

func TestDecodeMessage(t *testing.T) {
	var tests = []struct {
		name          string
		messageString string
		expected      *TestRecord
		expectedError string
	}{
		{
			name:          "empty messageString does not panic",
			messageString: "",
			expected:      nil,
			expectedError: "unexpected end of JSON input",
		},
		{
			name:          "garbage does not panic",
			messageString: "garbage",
			expected:      nil,
			expectedError: "illegal base64 data at input byte 4",
		},
		{
			name:          "encoded nil",
			messageString: "bnVsbA==",
			expected:      &TestRecord{},
			expectedError: "messageString is encoded <nil>",
		},
		{
			name: "as expected",
			messageString: "eyJJRCI6IjEyMzQ1IiwiTmFtZSI6Im5vbi10cml2aWFsIGVuY29k" +
				"ZSIsIkNyZWF0ZWQiOnt9LCJNb2RpZmllZCI6eyJzZWNvbmRzIjoxMzM3fX0=",
			expected: &TestRecord{
				ID:       "12345",
				Name:     "non-trivial encode",
				Created:  timestamppb.New(time.Unix(0, 0)),
				Modified: timestamppb.New(time.Unix(1337, 0)),
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert1.New(t)
			actual := new(TestRecord)
			actualError := DecodeMessage(tc.messageString, actual)
			if IsWhiteSpace(tc.expectedError) {
				assert.NoError(actualError)
				assert.Equal(tc.expected, actual)
			} else {
				assert.EqualError(actualError, tc.expectedError)
			}
		})
	}
}
