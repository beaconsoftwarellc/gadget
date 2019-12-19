package log

// Output represents a single output in a sequential chain of log message receivers.
type Output interface {
	// Level returns the LevelFlag that this logger output accepts
	Level() LevelFlag
	// Log the message passed
	Log(message Message)
}

type output struct {
	level LevelFlag
	log   func(Message)
}

func (o *output) Level() LevelFlag {
	return o.level
}

func (o *output) Log(message Message) {
	if nil != o.log {
		o.log(message)
	}
}

// NewOutput for use in the log chain.
func NewOutput(level LevelFlag, log func(Message)) Output {
	return &output{level: level, log: log}
}
