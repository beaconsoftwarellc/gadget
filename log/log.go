// Package log provides convenience functions to allow more granular levels of logging
// Requirements:
//   - structured logging - output will most likely be in json, but that shouldn't be known by the log creators
//   - log levels - limited as possible but convenient and impossible to screw up
//   - convenient setup and integration with standard baseLogger
//   - ability to include system/container/service names in logs by default
//   - include stack traces or other info
//
// Requirements:
//   - plain text output to local log file (io.Writer)
//   - structured json output to CloudWatch (or similar service)
//
//  inputs for each level:
//    error OR printf style
//
package log

import (
	"sync"
)

//NOTES: Be cautious regarding threadsafety and performance

// Level represents the importance of a message.
type Level string

// LevelFlag identifies the levels that a logger output accepts.
type LevelFlag uint8

const (
	// FlagDebug is the flag for enabling messages in a logger for Debug messages.
	FlagDebug, idxDebug = LevelFlag(1 << iota), iota
	// FlagAudit is the flag for enabling messages in a logger for Audit messages.
	FlagAudit, idxAudit
	// FlagInfo is the flag for enabling messages in a logger for Info messages.
	FlagInfo, idxInfo
	// FlagAccess is the flag for enabling messages in a logger for Access messages.
	FlagAccess, idxAccess
	// FlagWarn is the flag for enabling messages in a logger for Warn messages.
	FlagWarn, idxWarn
	// FlagError is the flag for enabling messages in a logger for Error messages.
	FlagError, idxError
	// FlagFatal is the flag for enabling messages in a logger for Fatal messages.
	FlagFatal, idxFatal
	// FlagAll has all log level flags set.
	FlagAll = (FlagFatal - 1) | FlagFatal
	// MaskDefault is set to only record Access, Error and Fatal messages.
	MaskDefault = FlagAccess | FlagError | FlagFatal
	// MaskVerbose filters everything not between Info and Fatal
	MaskVerbose = FlagInfo | FlagWarn | MaskDefault
	// MaskDebug filters nothing
	MaskDebug           = FlagAll
	standardStackOffset = 3
	// LevelFatal string for identifying message that are Fatal.
	LevelFatal Level = "FATAL"
	// LevelError string for identifying message that are Error.
	LevelError Level = "ERROR"
	// LevelWarn string for identifying message that are Warn.
	LevelWarn Level = "WARN"
	// LevelAudit string for identifying message that are Audit.
	LevelAudit Level = "AUDIT"
	// LevelInfo string for identifying message that are Info.
	LevelInfo Level = "INFO"
	// LevelAccess string for identifying message that are Access.
	LevelAccess Level = "ACCESS"
	// LevelDebug string for identifying message that are for Debug.
	LevelDebug Level = "DEBUG"
)

// Convert this level to a level flag
func (f Level) Convert() LevelFlag {
	var flag LevelFlag
	switch f {
	case LevelFatal:
		flag = FlagFatal
	case LevelError:
		flag = FlagError
	case LevelWarn:
		flag = FlagWarn
	case LevelAudit:
		flag = FlagAudit
	case LevelInfo:
		flag = FlagInfo
	case LevelAccess:
		flag = FlagAccess
	case LevelDebug:
		flag = FlagDebug
	}
	return flag
}

// Index for the passed level in the outputs array.
func (f Level) Index() (int, bool) {
	var idx LevelFlag = 0
	success := true
	switch f {
	case LevelFatal:
		idx = idxFatal
	case LevelError:
		idx = idxError
	case LevelWarn:
		idx = idxWarn
	case LevelAudit:
		idx = idxAudit
	case LevelInfo:
		idx = idxInfo
	case LevelAccess:
		idx = idxAccess
	case LevelDebug:
		idx = idxDebug
	default:
		success = false
	}
	return int(idx), success
}

type logger struct {
	identifier  string
	outputs     [][]Output
	stackOffset int
	mutex       sync.RWMutex
	sessionID   string
}

// New returns an implementation of the tiered logging interface
func New(id string, outputs ...Output) Logger {
	// initialize our output arrays
	array := make([][]Output, 7)
	for i := 0; i < len(array); i++ {
		array[i] = addOutput(uint(i), make([]Output, 0), outputs)
	}
	return &logger{identifier: id, outputs: array, stackOffset: standardStackOffset}
}

// New logger with a copied session and outputs as this logger. Changes to this logger
// will not affect the new logger.
func (l *logger) New(id string) Logger {
	return &logger{identifier: id,
		outputs:     l.outputs,
		stackOffset: standardStackOffset,
		sessionID:   l.sessionID,
	}
}

func (l *logger) SetSessionID(sessionID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.sessionID = sessionID
}

func (l *logger) GetSessionID() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.sessionID
}

func (l *logger) AddOutput(output Output) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for i := 0; i < len(l.outputs); i++ {
		l.outputs[i] = addOutput(uint(i), l.outputs[i], []Output{output})
	}
}

func addOutput(levelIdx uint, levelOutputs []Output, allOutputs []Output) []Output {
	for i := 0; i < len(allOutputs); i++ {
		if (1<<levelIdx)&allOutputs[i].Level() != 0 {
			levelOutputs = append(levelOutputs, allOutputs[i])
		}
	}
	return levelOutputs
}

// Fatal log message with the passed error
func (l *logger) Fatal(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelFatal, e))
	}
	return e
}

// Error log message with the passed error
func (l *logger) Error(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelError, e))
	}
	return e
}

// Warn log message with the passed error
func (l *logger) Warn(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelWarn, e))
	}
	return e
}

// Audit log message with the passed error
func (l *logger) Audit(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelAudit, e))
	}
	return e
}

// Info log message with the passed error
func (l *logger) Info(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelInfo, e))
	}
	return e
}

// Access log message with the passed error
func (l *logger) Access(e error) error {
	if nil != e {
		message := l.NewMessage(LevelAccess, e)
		message.Stack = nil
		l.log(message)
	}
	return e
}

// Debug log message with the passed error
func (l *logger) Debug(e error) error {
	if nil != e {
		l.log(l.NewMessage(LevelDebug, e))
	}
	return e
}

// Fatalf log entry with the passed format and arguments
func (l *logger) Fatalf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelFatal, format, args...)
	l.log(message)
	return message.Message
}

// Errorf log entry with the passed format and arguments
func (l *logger) Errorf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelError, format, args...)
	l.log(message)
	return message.Message
}

// Warnf log entry with the passed format and arguments
func (l *logger) Warnf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelWarn, format, args...)
	l.log(message)
	return message.Message
}

// Infof log entry with the passed format and arguments
func (l *logger) Infof(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelInfo, format, args...)
	l.log(message)
	return message.Message
}

// Accessf log entry with the passed format and arguments
func (l *logger) Accessf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelAccess, format, args...)
	message.Stack = nil
	l.log(message)
	return message.Message
}

// Debugf log entry with the passed format and arguments
func (l *logger) Debugf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelDebug, format, args...)
	l.log(message)
	return message.Message
}

// Auditf log entry with the passed format and arguments
func (l *logger) Auditf(format string, args ...interface{}) string {
	message := l.NewMessagef(LevelAudit, format, args...)
	l.log(message)
	return message.Message
}

func (l *logger) log(m *Message) {
	idx, ok := m.Level.Index()
	if ok {
		outputs := l.outputs[idx]
		for _, output := range outputs {
			output.Log(*m)
		}
	}
}
