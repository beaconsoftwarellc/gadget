package log

import (
	stdlog "log"
	"strings"

	"github.com/go-stack/stack"
)

// public logger - just assume we should write everything to the terminal until
// configured by a call to Global
var publicLogger = (New("", NewOutput(FlagAll, DefaultOutput))).(*logger)

// Global logger that is currently configured.
func Global() Logger {
	return publicLogger
}

// NewGlobal configures the global logger (analog of New()).  Also redirects the standard logger to output through the
// global logger.  This should be used as the first action in main().
func NewGlobal(id string, outputs ...Output) {
	stdlog.SetOutput(&stdLogAdapter{})
	//disable date/time, etc gathering from the standard logger
	stdlog.SetFlags(0)

	publicLogger = (New(id, outputs...)).(*logger)
	publicLogger.stackOffset++ //account for the extra call stack
}

// AddOutput to the global logger
func AddOutput(output Output) {
	publicLogger.AddOutput(output)
}

// Fatal logs a message regarding a failure that is severe enough to warrant process termination.
func Fatal(e error) error {
	return publicLogger.Fatal(e)
}

// Fatalf logs a message regarding a failure that is severe enough to warrant process termination.
func Fatalf(format string, a ...interface{}) string {
	return publicLogger.Fatalf(format, a...)
}

// Error logs a message regarding a non-fatal internal failure representing a probable code/system issue.
func Error(e error) error {
	return publicLogger.Error(e)
}

// Errorf logs a message regarding a non-fatal internal failure representing a probable code/system issue.
func Errorf(format string, a ...interface{}) string {
	return publicLogger.Errorf(format, a...)
}

// Warn logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
func Warn(e error) error {
	return publicLogger.Warn(e)
}

// Warnf logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
func Warnf(format string, a ...interface{}) string {
	return publicLogger.Warnf(format, a...)
}

// Audit logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
func Audit(e error) error {
	return publicLogger.Audit(e)
}

// Auditf logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
func Auditf(format string, a ...interface{}) string {
	return publicLogger.Auditf(format, a...)
}

// Info logs a message regarding debug information.
func Info(e error) error {
	return publicLogger.Info(e)
}

// Infof logs a message regarding debug information.
func Infof(format string, a ...interface{}) string {
	return publicLogger.Infof(format, a...)
}

// Access logs a message regarding access information.
func Access(e error) error {
	return publicLogger.Access(e)
}

// Accessf logs a message regarding access information.
func Accessf(format string, a ...interface{}) string {
	return publicLogger.Accessf(format, a...)
}

// Debug logs a message regarding debug information.
func Debug(e error) error {
	return publicLogger.Debug(e)
}

// Debugf logs a message regarding debug inforrmation.
func Debugf(format string, a ...interface{}) string {
	return publicLogger.Debugf(format, a...)
}

//////////////////////////
// Standard logger adapter
//////////////////////////

type stdLogAdapter struct{}

// Write exists to allow the standard logger to have a way to send log messages to this logging service.
// These messages will be recorded at INFO level.
func (*stdLogAdapter) Write(p []byte) (n int, err error) {
	message := publicLogger.NewMessagef(LevelInfo, "%s", strings.TrimSpace(string(p)))
	message.Caller = stack.Caller(standardStackOffset).String()
	publicLogger.log(message)
	return len(p), nil
}
