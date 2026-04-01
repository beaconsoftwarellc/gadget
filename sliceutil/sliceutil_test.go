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
