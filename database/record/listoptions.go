package record

// ListOptions provide limit and filtering capabilities for the List function
type ListOptions struct {
	Limit  uint
	Offset uint
}

// NewListOptions generates a ListOptions
func NewListOptions(limit int, offset int) ListOptions {
	return ListOptions{
		Limit:  uint(limit),
		Offset: uint(offset),
	}
}
