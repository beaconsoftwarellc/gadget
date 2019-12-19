package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-stack/stack"

	"github.com/beaconsoftwarellc/gadget/errors"
)

// Message encapsulates all fields that can possibly be a log message.
type Message struct {
	LogIdentifier string   `json:"LogIdentifier,omitempty"`
	SessionID     string   `json:"SessionID,omitempty"`
	Level         Level    `json:"Level,omitempty"`
	TimestampUnix int64    `json:"TimestampUnix,omitempty"`
	Timestamp     string   `json:"Timestamp,omitempty"`
	Caller        string   `json:"Caller,omitempty"`
	Message       string   `json:"Message,omitempty"`
	Stack         []string `json:"Stack,omitempty"`
	Error         error    `json:"Error,omitempty"`
}

// NewMessage with the passed log level and error
func (l *logger) NewMessage(level Level, err error) *Message {
	message := &Message{
		Message:   err.Error(),
		Error:     err,
		SessionID: l.sessionID,
	}
	message.setFields(level, l)
	return message
}

// NewMessagef with the passed level and message
func (l *logger) NewMessagef(level Level, format string, args ...interface{}) *Message {
	message := &Message{
		Message:   fmt.Sprintf(format, args...),
		SessionID: l.sessionID,
	}
	message.setFields(level, l)
	return message
}

func (m *Message) setFields(level Level, l *logger) {
	m.Level = level
	ts := time.Now().UTC()
	m.TimestampUnix = ts.Unix()
	m.Timestamp = ts.String()
	m.Caller = stack.Caller(l.stackOffset).String()
	m.LogIdentifier = l.identifier
	m.setTrace()
}

func (m *Message) setTrace() {
	var tracerErr errors.TracerError
	var ok bool
	if nil != m.Error {
		tracerErr, ok = m.Error.(errors.TracerError)
	}
	if ok {
		m.Stack = tracerErr.Trace()
	} else {
		m.Stack = getStack()
	}
}

// getStack retrieves the full stack (minus any runtime related functions) from where the log function was called
func getStack() []string {
	//may want to trim the log calls, may want to return this differently so that it can be formatted in other ways
	trace := stack.Trace().TrimRuntime()
	longTrace := make([]string, len(trace))
	for i, t := range trace {
		longTrace[i] = fmt.Sprintf("%+v", t)
	}
	return longTrace
}

// TTYString for logging this message to console.
func (m *Message) TTYString() string {
	if nil == m.Error {
		return fmt.Sprintf("[%s:%s] %s %s: %s (%s)\n", m.LogIdentifier, m.SessionID, m.Timestamp, m.Level, m.Message, m.Caller)
	}
	s := fmt.Sprintf("[%s:%s] %s %s: %+v (%s)\n", m.LogIdentifier, m.SessionID, m.Timestamp, m.Level, m.Error, m.Caller)
	if len(m.Stack) != 0 {
		s = fmt.Sprintf("%s%s\n", s, strings.Join(m.Stack, "\n"))
	}
	return s
}

// JSONString representation of this message.
func (m *Message) JSONString() string {
	b, err := json.Marshal(m)
	s := string(b)
	if nil != err {
		s = m.TTYString()
	}
	return s
}
