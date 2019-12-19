package log

import (
	stdlog "log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beaconsoftwarellc/gadget/errors"
	"github.com/beaconsoftwarellc/gadget/generator"
)

type ErrorPlus struct {
	A     string
	b     string
	trace []string
}

func NewErrorPlus(a, b string) *ErrorPlus {
	return &ErrorPlus{
		A:     a,
		b:     b,
		trace: errors.GetStackTrace(),
	}
}

func (err *ErrorPlus) Error() string {
	return err.b
}

// Trace returns the stack trace for the error
func (err *ErrorPlus) Trace() []string {
	return err.trace
}

func (err *ErrorPlus) String() string {
	return err.A
}

// Global instance

func TestStandardLogger(t *testing.T) {
	assert := assert.New(t)
	id := generator.String(5)
	var actual Message
	f := func(m Message) {
		actual = m
	}
	NewGlobal(id, NewOutput(FlagAll, f))
	msg := generator.String(7)
	stdlog.Print(msg)
	assert.True(strings.HasPrefix(actual.Caller, "log_test.go:"))
	assert.Equal(id, actual.LogIdentifier)
	assert.Equal(LevelInfo, actual.Level)
	assert.Equal(msg, actual.Message)
	assert.True(strings.HasPrefix(actual.Timestamp, "20"))
	assert.True(strings.HasSuffix(actual.Timestamp, "UTC"))
}

func TestLogFormat(t *testing.T) {
	assert := assert.New(t)
	var tests = []struct {
		function func(string, ...interface{}) string
		expected Level
	}{
		{
			function: Auditf,
			expected: LevelAudit,
		},
		{
			function: Infof,
			expected: LevelInfo,
		},
		{
			function: Accessf,
			expected: LevelAccess,
		},
		{
			function: Warnf,
			expected: LevelWarn,
		},
		{
			function: Errorf,
			expected: LevelError,
		},
		{
			function: Debugf,
			expected: LevelDebug,
		},
	}
	for _, t := range tests {
		id := generator.String(5)
		var actual Message
		f := func(m Message) {
			actual = m
		}
		NewGlobal(id, NewOutput(FlagAll, f))
		msg := generator.String(7)
		t.function(msg)
		assert.True(strings.HasPrefix(actual.Caller, "log_test.go"), "test for %s failed", t.expected)
		assert.Equal(id, actual.LogIdentifier, id)
		assert.Equal(t.expected, actual.Level, t.expected)
		assert.Equal(msg, actual.Message)
		assert.True(strings.HasPrefix(actual.Timestamp, "20"), t.expected)
		assert.True(strings.HasSuffix(actual.Timestamp, "UTC"), t.expected)
	}
}

func TestLogErrorObj(t *testing.T) {
	assert := assert.New(t)
	var tests = []struct {
		function func(err error) error
		expected Level
		noCaller bool
	}{
		{
			function: Audit,
			expected: LevelAudit,
			noCaller: true,
		},
		{
			function: Info,
			expected: LevelInfo,
		},
		{
			function: Access,
			expected: LevelAccess,
		},
		{
			function: Warn,
			expected: LevelWarn,
		},
		{
			function: Error,
			expected: LevelError,
		},
		{
			function: Fatal,
			expected: LevelFatal,
		},
		{
			function: Debug,
			expected: LevelDebug,
		},
	}
	for _, t := range tests {
		id := generator.String(5)
		var actual Message
		f := func(m Message) {
			actual = m
		}
		NewGlobal(id, NewOutput(FlagAll, f))
		err := errors.New(generator.String(7))
		t.function(err)
		if !t.noCaller {
			assert.True(strings.HasPrefix(actual.Caller, "log_test.go"), "test for %s failed", t.expected)
		}

		assert.Equal(id, actual.LogIdentifier, id)
		assert.Equal(t.expected, actual.Level, t.expected)
		assert.Equal(err.Error(), actual.Message)
		assert.True(strings.HasPrefix(actual.Timestamp, "20"), t.expected)
		assert.True(strings.HasSuffix(actual.Timestamp, "UTC"), t.expected)
	}
}

func TestLogger_AddOutput(t *testing.T) {
	assert := assert.New(t)
	var actual Message
	var actual1 Message
	f := func(m Message) {
		actual = m
	}
	NewGlobal("TestLogger_AddOutput", NewOutput(FlagAll, f))

	f1 := func(m Message) {
		actual1 = m
	}
	output := NewOutput(FlagError, f1)
	AddOutput(output)

	assert.Contains(publicLogger.outputs[idxError], output)
	expected := generator.String(10)
	Warnf(expected)
	assert.Equal(expected, actual.Message)
	assert.Empty(actual1.Message)

	expected = generator.String(10)
	Errorf(expected)
	assert.Equal(expected, actual.Message)
	assert.Equal(expected, actual1.Message)
}
