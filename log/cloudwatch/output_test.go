package cloudwatch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/log"
)

func Test_EnsureGroupNameIsValid(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Test Emtpy Input",
			input:    "",
			expected: "EmptyLogGroupName",
		},
		{
			name:     "Test Safe Input",
			input:    "asdf_-/.",
			expected: "asdf_-/.",
		},
		{
			name:     "Test Emoji Input",
			input:    "asdf😜",
			expected: "asdf",
		},
		{
			name:     "Test results in whitespace",
			input:    "😜",
			expected: "EmptyLogGroupName",
		},
	}
	for _, test := range tests {
		actual := EnsureGroupNameIsValid(test.input)
		assert.Equal(test.expected, actual, "(%s) EnsureGroupNameIsValid('%s') = '%s', Expected '%s'",
			test.name, test.input, actual, test.expected)
	}
}

func Test_EnsureStreamNameIsValid(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Test Emtpy Input",
			input:    "",
			expected: "EmptyLogStreamName",
		},
		{
			name:     "Test Safe Input",
			input:    "asdf_-/.",
			expected: "asdf_-/.",
		},
		{
			name:     "Test Emoji Input",
			input:    "asdf😜",
			expected: "asdf😜",
		},
		{
			name:     "Test results in whitespace",
			input:    "*:",
			expected: "EmptyLogStreamName",
		},
	}
	for _, test := range tests {
		actual := EnsureStreamNameIsValid(test.input)
		assert.Equal(test.expected, actual, "(%s) EnsureGroupNameIsValid('%s') = '%s', Expected '%s'",
			test.name, test.input, actual, test.expected)
	}
}

func Test_Skip(t *testing.T) {
	// we are not testing the actual integration with cloudwatch since the API is too big to replicate
	// and the tests would be testing the framework.
	// Run this if you make changes and check that the messages make it
	// into cloudwatch.
	t.SkipNow()
	assert := assert.New(t)
	admin, err := GetAdministration()
	if !assert.NoError(err) {
		return
	}
	output, err := admin.GetOutput("testGroup", "testStream", log.FlagAll)
	if !assert.NoError(err) {
		return
	}
	message := log.Message{Message: "some riot text", TimestampUnix: time.Now().UTC().Unix()}
	output.Log(message)
	message = log.Message{Message: "some riot text", TimestampUnix: time.Now().UTC().Unix()}
	output.Log(message)
	message = log.Message{Message: "some riot text", TimestampUnix: time.Now().UTC().Unix()}
	output.Log(message)
	time.Sleep(1)
	assert.Fail("adsf")
}
