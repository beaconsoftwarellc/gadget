package collection

import (
	"fmt"
	"reflect"
	"testing"
)

type args[T comparable] struct {
	a Set[T]
	b Set[T]
}

type testCase[T comparable] struct {
	args args[T]
	want Set[T]
}

func TestUnion(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		testCases := []testCase[int]{
			{
				args: args[int]{a: NewSet(1, 2), b: NewSet(2, 3, 4)},
				want: NewSet(1, 2, 3, 4),
			},
			{
				args: args[int]{a: NewSet[int](), b: NewSet(1)},
				want: NewSet(1),
			},
			{
				args: args[int]{a: NewSet[int](), b: NewSet[int]()},
				want: NewSet[int](),
			},
		}

		runUnionTests(t, testCases)
	})

	t.Run("strings", func(t *testing.T) {
		testCases := []testCase[string]{
			{
				args: args[string]{a: NewSet("1", "2"), b: NewSet("2", "3", "4")},
				want: NewSet("1", "2", "3", "4"),
			},
			{
				args: args[string]{a: NewSet("b"), b: NewSet("b")},
				want: NewSet("b"),
			},
			{
				args: args[string]{a: NewSet[string](), b: NewSet[string]()},
				want: NewSet[string](),
			},
		}

		runUnionTests(t, testCases)
	})

}

func runUnionTests[T comparable](t *testing.T, tcs []testCase[T]) {
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			if got := Union(tc.args.a, tc.args.b); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Union() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDisjunction(t *testing.T) {
	type args struct {
		a Set[string]
		b Set[string]
	}
	tests := []struct {
		name string
		args args
		want Set[string]
	}{
		{
			args: args{a: NewSet("a"), b: NewSet("a")},
			want: NewSet[string](),
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
		a Set[string]
		b Set[string]
	}
	tests := []struct {
		name string
		args args
		want Set[string]
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
		ms   Set[string]
		want int
	}{
		{
			ms:   NewSet[string](),
			want: 0,
		},
		{
			ms:   NewSet("a"),
			want: 1,
		},
		{
			ms:   NewSet("1", "2"),
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
