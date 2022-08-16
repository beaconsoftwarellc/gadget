package binutil

import (
	"golang.org/x/exp/constraints"
)

// GetDiscreteStatuses in a slice of valid statuses from the given T value
func GetDiscreteStatuses[T constraints.Integer](s T, max T) []T {
	var (
		resp = []T{}
		i    T
	)

	for i = 1; i < max; i = i << 1 {
		if i&s == i {
			resp = append(resp, i)
		}
	}

	return resp
}
