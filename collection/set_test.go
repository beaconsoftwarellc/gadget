package collection

import (
	"reflect"
	"testing"
)

func TestUnion(t *testing.T) {
	type args struct {
		a Set
		b Set
	}
	tests := []struct {
		name string
		args args
		want Set
	}{
		{
			args: args{a: NewSet(), b: NewSet()},
			want: NewSet(),
		},
		{
			args: args{a: NewSet(1, 2), b: NewSet(2, 3, 4)},
			want: NewSet(1, 2, 3, 4),
		},
		{
			args: args{a: NewSet("a"), b: NewSet()},
			want: NewSet("a"),
		},
		{
			args: args{a: NewSet(), b: NewSet(1)},
			want: NewSet(1),
		},
		{
			args: args{a: NewSet(1), b: NewSet("a")},
			want: NewSet(1, "a"),
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
		a Set
		b Set
	}
	tests := []struct {
		name string
		args args
		want Set
	}{
		{
			args: args{a: NewSet("a"), b: NewSet("a")},
			want: NewSet(),
		},
		{
			args: args{a: NewSet("a", "b"), b: NewSet("a", "b", "c")},
			want: NewSet("c"),
		},
		{
			args: args{a: NewSet("a", "b", "c"), b: NewSet("a", "b")},
			want: NewSet("c"),
		},
		{
			args: args{a: NewSet("a", "b", "c"), b: NewSet("a", "b", "d")},
			want: NewSet("c", "d"),
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
		a Set
		b Set
	}
	tests := []struct {
		name string
		args args
		want Set
	}{
		{
			args: args{a: NewSet("a"), b: NewSet("a")},
			want: NewSet("a"),
		},
		{
			args: args{a: NewSet("a", "b"), b: NewSet("a", "b", "c")},
			want: NewSet("a", "b"),
		},
		{
			args: args{a: NewSet("a", "b", "c"), b: NewSet("a", "b")},
			want: NewSet("a", "b"),
		},
		{
			args: args{a: NewSet("a", "b", "c"), b: NewSet("a", "b", "d")},
			want: NewSet("a", "b"),
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
		ms   Set
		want int
	}{
		{
			ms:   NewSet(),
			want: 0,
		},
		{
			ms:   NewSet("a"),
			want: 1,
		},
		{
			ms:   NewSet(1, 2),
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
