package fileutil

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/beaconsoftwarellc/gadget/generator"
)

// FileExists exists and is accessible and the specified path.
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	r := true
	if os.IsNotExist(err) || stat.IsDir() {
		r = false
	}
	return r
}

// EnsureDir at the specified path with the specified mode.
func EnsureDir(path string, mode os.FileMode) (os.FileInfo, error) {
	// MkdirAll returns nil if the directory already exists.
	err := os.MkdirAll(path, mode)
	if nil != err {
		return nil, err
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Mode() != mode {
		os.Chmod(path, mode)
		// need to restat to reflect mode change.
		fi, err = os.Stat(path)
	}
	return fi, err
}

// TempFile with the passed contents.
func TempFile(contents string) (string, error) {
	dir := path.Join("/tmp", "orient")
	file := path.Join(dir, generator.String(10)+".tmp")
	_, err := EnsureDir(dir, 0777)
	if nil != err {
		return file, err
	}
	f, err := os.Create(file)
	if nil != err {
		return file, err
	}
	_, err = f.Write([]byte(contents))
	return file, err
}

// ReadLines from a file and return them including any errors.
func ReadLines(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	lines := []string{}
	if nil != err {
		return lines, err
	}
	reader := bufio.NewReader(file)
	prefix := ""
	for nil == err {
		bytes, isPrefix, err := reader.ReadLine()
		if nil != err {
			break
		}
		line := string(bytes)
		if isPrefix {
			prefix = prefix + line
		} else {
			lines = append(lines, prefix+line)
			prefix = ""
		}
	}
	file.Close()
	return lines, err
}

// DownloadToMemory downloads a file from a url to memory
func DownloadToMemory(url string) ([]byte, error) {
	var contents []byte
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}
