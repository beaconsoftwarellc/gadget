package fileutil

import (
	"bytes"
	"io"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// ReadYamlFromFile at path filename into the target interface.
func ReadYamlFromFile(filename string, target interface{}) error {
	data, err := os.ReadFile(filename)
	if nil != err {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// WriteYamlToFile at path filename sourcing the data from the passed target.
func WriteYamlToFile(filename string, target interface{}) error {
	data, err := yaml.Marshal(target)
	if nil != err {
		return err
	}
	return os.WriteFile(filename, data, 0777)
}

// WriteYamlToWriter returning any errors that occur.
func WriteYamlToWriter(writer io.Writer, target interface{}) error {
	data, err := yaml.Marshal(target)
	buffer := bytes.NewBuffer(data)
	if nil != err {
		return err
	}
	_, err = io.Copy(writer, buffer)
	return err
}
