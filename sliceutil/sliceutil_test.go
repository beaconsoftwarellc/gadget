package sliceutil

import (
	"errors"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name        string
		source      func(qb.LimitOffset) ([]int, int, error)
		expected    []int
		expectError bool
	}{
		{
			name: "simple case",
			source: func(qb.LimitOffset) ([]int, int, error) {
				return []int{1, 2, 3}, 3, nil
			},
			expected:    []int{1, 2, 3},
			expectError: false,
		},
		{
			name: "multiple batches",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				batches := [][]int{{1, 2}, {3, 4}, {5}}
				index := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					if index >= len(batches) {
						return nil, 5, nil
					}
					res := batches[index]
					index++
					return res, 5, nil
				}
			}(),
			expected:    []int{1, 2, 3, 4, 5},
			expectError: false,
		},
		{
			name: "empty batches",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				return func(qb.LimitOffset) ([]int, int, error) {
					return nil, 0, nil
				}
			}(),
			expected:    []int{},
			expectError: false,
		},
		{
			name: "error at the outset",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				return func(qb.LimitOffset) ([]int, int, error) {
					return nil, 0, errors.New("fetch error")
				}
			}(),
			expected:    []int{},
			expectError: true,
		},
		{
			name: "batch with error mid-batch",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				batches := [][]int{{1, 2}, nil, {3, 4}}
				errorsList := []error{nil, errors.New("fetch error"), nil}
				index := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					res := batches[index]
					err := errorsList[index]
					index++
					return res, 4, err
				}
			}(),
			expected:    []int{1, 2},
			expectError: true,
		},
		{
			name: "infinite data source",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				count := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					count++
					return []int{count}, 6, nil
				}
			}(),
			expected:    []int{1, 2, 3, 4, 5, 6},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []int
			var caughtError error

			seq := Flatten(qb.NewLimitOffset[int]().
				SetLimit(5).SetOffset(0), tt.source)
			for item, err := range seq {
				if err != nil {
					caughtError = err
					break
				}
				result = append(result, item)
			}

			if caughtError != nil && !tt.expectError {
				t.Errorf("unexpected error: %v", caughtError)
			} else if caughtError == nil && tt.expectError {
				t.Error("expected an error but got none")
			}

			if tt.expected == nil {
				if len(result) > 0 {
					t.Errorf("expected no items, got %v", result)
				}
			} else {
				for i, v := range tt.expected {
					if i >= len(result) || result[i] != v {
						t.Errorf("at index %d, expected %v, got %v", i, v, result)
					}
				}
				if len(result) > len(tt.expected) {
					t.Errorf("unexpected extra elements: %v", result[len(tt.expected):])
				}
			}
		})
	}
}

func TestExhaust(t *testing.T) {
	tests := []struct {
		name        string
		source      func(qb.LimitOffset) ([]int, int, error)
		expected    []int
		expectError bool
	}{
		{
			name: "simple case",
			source: func(qb.LimitOffset) ([]int, int, error) {
				return []int{1, 2, 3}, 3, nil
			},
			expected:    []int{1, 2, 3},
			expectError: false,
		},
		{
			name: "multiple batches",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				batches := [][]int{{1, 2}, {3, 4}, {5}}
				index := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					if index >= len(batches) {
						return nil, 5, nil
					}
					res := batches[index]
					index++
					return res, 5, nil
				}
			}(),
			expected:    []int{1, 2, 3, 4, 5},
			expectError: false,
		},
		{
			name: "empty result",
			source: func(qb.LimitOffset) ([]int, int, error) {
				return nil, 0, nil
			},
			expected:    []int{},
			expectError: false,
		},
		{
			name: "error on first batch",
			source: func(qb.LimitOffset) ([]int, int, error) {
				return nil, 0, errors.New("fetch error")
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "error after some batches",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				batches := [][]int{{1, 2}, nil}
				errorsList := []error{nil, errors.New("fetch error")}
				index := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					res := batches[index]
					err := errorsList[index]
					index++
					return res, 3, err
				}
			}(),
			expected:    nil,
			expectError: true,
		},
		{
			name: "single empty batch",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				called := false
				return func(qb.LimitOffset) ([]int, int, error) {
					if called {
						return nil, 0, nil
					}
					called = true
					return []int{}, 0, nil
				}
			}(),
			expected:    []int{},
			expectError: false,
		},
		{
			name: "large dataset",
			source: func() func(qb.LimitOffset) ([]int, int, error) {
				batches := make([][]int, 10)
				for i := 0; i < 10; i++ {
					batches[i] = []int{i*10 + 1, i*10 + 2, i*10 + 3}
				}
				index := 0
				return func(qb.LimitOffset) ([]int, int, error) {
					if index >= len(batches) {
						return nil, 30, nil
					}
					res := batches[index]
					index++
					return res, 30, nil
				}
			}(),
			expected:    []int{1, 2, 3, 11, 12, 13, 21, 22, 23, 31, 32, 33, 41, 42, 43, 51, 52, 53, 61, 62, 63, 71, 72, 73, 81, 82, 83, 91, 92, 93},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Exhaust(tt.source)

			if err != nil && !tt.expectError {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil && tt.expectError {
				t.Error("expected an error")
			}

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if len(result) != len(tt.expected) {
					t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				}
				for i, v := range tt.expected {
					if i >= len(result) || result[i] != v {
						t.Errorf("at index %d, expected %v, got %v", i, v, result[i])
					}
				}
			}
		})
	}
}
