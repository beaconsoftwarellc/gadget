//go:generate mockgen -source=$GOFILE -package $GOPACKAGE -destination interface.mock.gen.go

package log

// Logger is the tiered level logging interface
type Logger interface {
	// New logger with a copied session and outputs as this logger. Changes to this logger
	// will not affect the new logger.
	New(id string) Logger
	// GetSessionID that is currently in use on this logger
	GetSessionID() string
	// SetSessionID that is currently in use on this logger
	SetSessionID(string)
	// Fatal logs a message regarding a failure that is severe enough to warrant process termination.
	Fatal(e error) error
	// Fatalf logs a message regarding a failure that is severe enough to warrant process termination.
	Fatalf(format string, a ...interface{}) string
	// Error logs a message regarding a non-fatal internal failure representing a probable code/system issue.
	Error(e error) error
	// Errorf logs a message regarding a non-fatal internal failure representing a probable code/system issue.
	Errorf(format string, a ...interface{}) string
	// Warn logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
	Warn(e error) error
	// Warnf logs a message regarding a failure type that is expected to occasionally occur (ex: invalid arguments).
	Warnf(format string, a ...interface{}) string
	// Audit logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
	Audit(e error) error
	// Auditf logs a message regarding normal events that are to be systematically recorded (ex: RMS messages).
	Auditf(format string, a ...interface{}) string
	// Access logs a message regarding information about the state of a successfully executing system.
	Access(e error) error
	// Accessf logs a message regarding information about the state of a successfully executing system.
	Accessf(format string, a ...interface{}) string
	// Info logs a message regarding information about the state of a successfully executing system.
	Info(e error) error
	// Infof logs a message regarding information about the state of a successfully executing system.
	Infof(format string, a ...interface{}) string
	// Debug logs a message regarding debug information that would be overly verbose for an info message.
	Debug(e error) error
	// Debugf logs a message regarding debug information that would be overly verbose for an info message.
	Debugf(format string, a ...interface{}) string
	// AddOutput to this logger
	AddOutput(Output)
}

// Tracer provides a stack trace
type Tracer interface {
	Trace() []string
}

// TracerError is an amalgamation of the Tracer and error interfaces
type TracerError interface {
	Tracer
	error
}
