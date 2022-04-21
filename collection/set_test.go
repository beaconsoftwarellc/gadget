package collection

import (
	"reflect"
	"testing"
)

func TestUnion(t *testing.T) {
	type args struct {
		a Set[interface{}]
		b Set[interface{}]
	}
	tests := []struct {
		name string
		args args
		want Set[interface{}]
	}{
		{
			args: args{a: NewSet[interface{}](), b: NewSet[interface{}]()},
			want: NewSet[interface{}](),
		},
		{
			args: args{a: NewSet[interface{}](1, 2), b: NewSet[interface{}](2, 3, 4)},
			want: NewSet[interface{}](1, 2, 3, 4),
		},
		{
			args: args{a: NewSet[interface{}]("a"), b: NewSet[interface{}]()},
			want: NewSet[interface{}]("a"),
		},
		{
			args: args{a: NewSet[interface{}](), b: NewSet[interface{}](1)},
			want: NewSet[interface{}](1),
		},
		{
			args: args{a: NewSet[interface{}](1), b: NewSet[interface{}]("a")},
			want: NewSet[interface{}](1, "a"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Union(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisjunction(t *testing.T) {
	type args struct {
		a Set[interface{}]
		b Set[interface{}]
	}
	tests := []struct {
		name string
		args args
		want Set[interface{}]
	}{
		{
			args: args{a: NewSet[interface{}]("a"), b: NewSet[interface{}]("a")},
			want: NewSet[interface{}](),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b"), b: NewSet[interface{}]("a", "b", "c")},
			want: NewSet[interface{}]("c"),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b", "c"), b: NewSet[interface{}]("a", "b")},
			want: NewSet[interface{}]("c"),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b", "c"), b: NewSet[interface{}]("a", "b", "d")},
			want: NewSet[interface{}]("c", "d"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Disjunction(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Disjunction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	type args struct {
		a Set[interface{}]
		b Set[interface{}]
	}
	tests := []struct {
		name string
		args args
		want Set[interface{}]
	}{
		{
			args: args{a: NewSet[interface{}]("a"), b: NewSet[interface{}]("a")},
			want: NewSet[interface{}]("a"),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b"), b: NewSet[interface{}]("a", "b", "c")},
			want: NewSet[interface{}]("a", "b"),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b", "c"), b: NewSet[interface{}]("a", "b")},
			want: NewSet[interface{}]("a", "b"),
		},
		{
			args: args{a: NewSet[interface{}]("a", "b", "c"), b: NewSet[interface{}]("a", "b", "d")},
			want: NewSet[interface{}]("a", "b"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Intersection(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Intersection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapSet_Size(t *testing.T) {
	tests := []struct {
		name string
		ms   Set[interface{}]
		want int
	}{
		{
			ms:   NewSet[interface{}](),
			want: 0,
		},
		{
			ms:   NewSet[interface{}]("a"),
			want: 1,
		},
		{
			ms:   NewSet[interface{}](1, 2),
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ms.Size(); got != tt.want {
				t.Errorf("mapSet.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}
