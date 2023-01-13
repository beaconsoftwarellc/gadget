package record

import "golang.org/x/exp/constraints"

// ListOptions provide limit and filtering capabilities for the List function
type ListOptions struct {
	Limit  uint
	Offset uint
}

// NewListOptions generates a ListOptions
func NewListOptions[T constraints.Integer](limit T, offset T) *ListOptions {
	return &ListOptions{
		Limit:  uint(limit),
		Offset: uint(offset),
	}
}
