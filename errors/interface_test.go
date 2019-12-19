package errors

import (
	"reflect"
	"testing"
)

func Test_getStackTrace(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{want: []string{
			"github.com/beaconsoftwarellc/gadget/errors/interface_test.go:19",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStackTrace(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStackTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
		want TracerError
	}{
		{
			args: args{message: "foo"},
			want: &errorTracer{
				message: "foo",
				trace: []string{
					"github.com/beaconsoftwarellc/gadget/errors/interface_test.go:57",
				},
			},
		},
		{
			args: args{message: "foo %s %s", args: []interface{}{"a", "b"}},
			want: &errorTracer{
				message: "foo a b",
				trace: []string{
					"github.com/beaconsoftwarellc/gadget/errors/interface_test.go:57",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.message, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
