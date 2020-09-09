package log

import (
	"fmt"
	"os"
	"sync"
)

type fileOutput struct {
	level LevelFlag
	file *os.File
	filepath string
	mutex sync.Mutex
}

// NewFileOutput that writes messages of the passed level to the passed file path.
func NewFileOutput(level LevelFlag, path string) (Output, error) {
	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if nil != err {
		return nil, err
	}
	return &fileOutput{
		level:    level,
		file: 		fd,
		filepath: path,
	}, nil
}

func (o *fileOutput) Level() LevelFlag {
	return o.level
}

func (o *fileOutput) Log(message Message) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	_, err := o.file.WriteString(message.TTYString())
	if nil != err {
		fmt.Printf("failed to write to log file '%s': %s\n", o.filepath, err)
	}
}

