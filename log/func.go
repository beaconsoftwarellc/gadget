package log

import (
	"os"
	"strconv"
)

const (
	// awsLogs indicates use AWS logging if set, else log to default
	awsLogs = "AWS_LOGS"
	// MaskEnv is the environment variable that stores the logging mask
	MaskEnv = "LOGGING_MASK"
)

// FunctionFromEnv reads the environment variables and returns an appropriate log function
func FunctionFromEnv() Output {
	logLevel := loggingMaskFromEnv()
	_, set := os.LookupEnv(awsLogs)
	if set {
		return NewOutput(logLevel, JSONOutput)
	}
	return NewOutput(logLevel, DefaultOutput)
}

func loggingMaskFromEnv() LevelFlag {
	logLevelStr, ok := os.LookupEnv(MaskEnv)
	if ok {
		logLevelInt, err := strconv.Atoi(logLevelStr)
		ok = nil == err
		if ok {
			return LevelFlag(logLevelInt)
		}
	}
	return MaskDefault
}

// DefaultOutput mimics the behavior of the default logging package by printing to StdErr.
// This is intended to be passed as one of the output functions to New() or Global().
func DefaultOutput(m Message) {
	flag := m.Level.Convert()
	if flag&(FlagInfo|FlagAudit|FlagWarn) > 0 {
		m.Stack = []string{}
	}
	os.Stderr.Write([]byte(m.TTYString()))
}

// JSONOutput formats the log map into Json and then outputs to stderr (appropriate for aws production services)
func JSONOutput(m Message) {
	os.Stderr.Write([]byte(m.JSONString() + "\n"))
}

// ExitOnError will exit if an error is returned
func ExitOnError(err error) {
	if nil != err {
		Error(err)
		os.Exit(1)
	}
}
