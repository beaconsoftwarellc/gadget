package fileutil

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/generator"
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
	err := os.MkdirAll(testPath, 0777)
	if err != nil {
		t.Error(err)
	}
	testfile := path.Join(testPath, "test.txt")
	_, err = os.Create(testfile)
	if err != nil {
		t.Error(err)
	}
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
			want: true,
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

func TestRemoveFileMatches(t *testing.T) {
	testPath := getTestPath()
	err := os.MkdirAll(testPath, 0777)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name          string
		files         []string
		pattern       string
		wantDeleted   []string
		wantRemaining []string
	}{
		{
			name:          "empty matching pattern",
			files:         []string{"one.txt", "two.txt"},
			pattern:       "",
			wantDeleted:   []string{},
			wantRemaining: []string{"one.txt", "two.txt"},
		},
		{
			name:          "no matches",
			files:         []string{"one.txt", "two.txt"},
			pattern:       path.Join(testPath, "no-matches"),
			wantDeleted:   []string{},
			wantRemaining: []string{"one.txt", "two.txt"},
		},
		{
			name:          "match",
			files:         []string{"one.txt", "two.txt"},
			pattern:       path.Join(testPath, "one.txt"),
			wantDeleted:   []string{"one.txt"},
			wantRemaining: []string{"two.txt"},
		},
		{
			name:          "wildcard in pattern",
			files:         []string{"one.txt", "two.txt"},
			pattern:       path.Join(testPath, "one*"),
			wantDeleted:   []string{"one.txt"},
			wantRemaining: []string{"two.txt"},
		},
		{
			name:          "delete all in dir",
			files:         []string{"one.txt", "two.txt"},
			pattern:       path.Join(testPath, "*"),
			wantDeleted:   []string{"one.txt", "two.txt"},
			wantRemaining: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.files {
				_, err = os.Create(path.Join(testPath, f))
				if err != nil {
					t.Error(err)
				}
			}

			err = RemoveFileMatches(tt.pattern)
			if err != nil {
				t.Error(err)
			}

			for _, f := range tt.wantDeleted {
				if FileExists(path.Join(testPath, f)) {
					t.Errorf("RemoveFileMatches(\"%s\") = unexpected file not deleted %v", tt.pattern, f)
				}
			}

			for _, f := range tt.wantRemaining {
				if !FileExists(path.Join(testPath, f)) {
					t.Errorf("RemoveFileMatches(\"%s\") = unexpected file deleted %v", tt.pattern, f)
				}
			}
		})
	}
}
