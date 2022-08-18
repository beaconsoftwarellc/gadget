package database

const (
	// DefaultMaxTries documentation hur
	DefaultMaxTries = 10
	// DefaultMaxLimit for row counts on select queries
	DefaultMaxLimit = 100
)

// ListOptions provide limit and filtering capabilities for the List function
type ListOptions struct {
	Limit  uint
	Offset uint
}

// NewListOptions generates a ListOptions
func NewListOptions(limit uint, offset uint) *ListOptions {
	return &ListOptions{
		Limit:  limit,
		Offset: offset,
	}
}

// Model represents the basic table information from a DB Record
type Model struct {
	Name         string
	PrimaryKey   string
	ReadColumns  []string
	WriteColumns []string
}
