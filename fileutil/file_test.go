package fileutil

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/beaconsoftwarellc/gadget/generator"
)

func getTestPath() string {
	uuid := generator.String(10)
	rootPath := fmt.Sprintf("/tmp/%s/%s", "fileutil", strings.TrimSpace(uuid))
	return rootPath
}

func TestEnsureDirectory(t *testing.T) {
	// test happy path
	dirname := getTestPath()
	_, err := EnsureDir(dirname, 0777)
	if err != nil {
		t.Error(err)
	}
	err = os.Remove(dirname)
	if err != nil {
		t.Error(err)
	}

	// test creating a path when a subdirectory is a file
	f, err := os.Create(dirname)
	if err != nil {
		t.Error(err)
	}
	_, err = f.WriteString("testensure")
	if err != nil {
		t.Error(err)
	}
	err = f.Close()
	if err != nil {
		t.Error(err)
	}

	_, err = EnsureDir(dirname, 0777)
	if err == nil {
		t.Error("Existing file with dirname should fail.")
	}

	_, err = EnsureDir(dirname+"/foo", 0777)
	if err == nil {
		t.Error("File in subtree should fail.")
	}
	os.Remove(dirname)
}

func TestFileExists(t *testing.T) {
	type args struct {
		path string
	}
	testPath := getTestPath()
	testfile := path.Join(testPath, "test.txt")
	os.Create(testfile)
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "dir false",
			args: args{path: "/tmp/"},
			want: false,
		},
		{
			name: "empty false",
			args: args{path: ""},
			want: false,
		},
		{
			name: "file true",
			args: args{path: testfile},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.path); got != tt.want {
				t.Errorf("FileExists(\"%s\") = %v, want %v", tt.args.path, got, tt.want)
			}
		})
	}
}
