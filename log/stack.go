package log

import (
	"fmt"

	"github.com/beaconsoftwarellc/gadget/collection"
)

// StackLogger implments the tiered logging interface with a string stack for the messages
type StackLogger struct {
	messages collection.StringStack
}

// New returns this logger
func (l *StackLogger) New(id string) Logger {
	return l
}

// GetSessionID empty string
func (l *StackLogger) GetSessionID() string {
	return ""
}

// SetSessionID noop
func (l *StackLogger) SetSessionID(id string) {

}

// NewStackLogger puts the log messages onto a stack
func NewStackLogger() *StackLogger {
	return &StackLogger{messages: collection.NewStringStack()}
}

// AddOutput StackLogger does not support this for obvious reasons
func (l *StackLogger) AddOutput(output Output) {
	// StackLogger does not support this for obvious reasons
}

// IsEmpty checks if the stack has no messages
func (l *StackLogger) IsEmpty() bool {
	return 0 == l.messages.Size()
}

// Pop returns the last message from the message stack
func (l *StackLogger) Pop() (string, error) {
	return l.messages.Pop()
}

// Fatal logs a message regarding a failure that is severe enough to warrant process termination.
func (l *StackLogger) Fatal(e error) error {
	return l.push(e)
}

func (l *StackLogger) push(err error) error {
	if nil == err {
		return err
	}
	l.messages.Push(err.Error())
	return err
}

func (l *StackLogger) pushf(format string, a ...interface{}) string {
	s := fmt.Sprintf(format, a...)
	l.messages.Push(s)
	return s
}

// Fatalf logs a message regarding a failure that is severe enough to warrant process termination.
func (l *StackLogger) Fatalf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Error logs a message regarding a non-fatal internal failure representing a probable code/system issue.
func (l *StackLogger) Error(e error) error {
	return l.push(e)
}

// Errorf logs a message regarding a non-fatal internal failure representing a probable code/system issue.
func (l *StackLogger) Errorf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Warn logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
func (l *StackLogger) Warn(e error) error {
	return l.push(e)
}

// Warnf logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
func (l *StackLogger) Warnf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Audit logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
func (l *StackLogger) Audit(e error) error {
	return l.push(e)
}

// Auditf logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
func (l *StackLogger) Auditf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Info logs a message regarding debug information.
func (l *StackLogger) Info(e error) error {
	return l.push(e)
}

// Infof logs a message regarding debug information.
func (l *StackLogger) Infof(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Access logs a message regarding access information.
func (l *StackLogger) Access(e error) error {
	return l.push(e)
}

// Accessf logs a message regarding access information.
func (l *StackLogger) Accessf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}

// Debug logs a message regarding debug information.
func (l *StackLogger) Debug(e error) error {
	return l.push(e)
}

// Debugf logs a message regarding debug information.
func (l *StackLogger) Debugf(format string, a ...interface{}) string {
	return l.pushf(format, a...)
}
