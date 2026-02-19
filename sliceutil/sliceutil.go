package sliceutil

import (
	"iter"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
)

// Flatten is a utility function for flattening a loop that ranges over
// an array of an array of items returned from a function. Useful in paging
// queries.
func Flatten[T any](limitOffset qb.LimitOffset, source func(limitOffset qb.LimitOffset) ([]T, int, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		setter := qb.NewLimitOffset[uint]().SetOffset(limitOffset.Offset()).SetLimit(limitOffset.Limit())
		for items, total, err := source(setter); setter.Offset() < uint(total); setter.SetOffset(setter.Offset() + uint(len(items))) {
			var (
				item T
				c    bool
			)
			if err != nil {
				c = yield(item, err)
			}
			for _, item := range items {
				c = yield(item, nil)
			}
			if !c {
				return
			}
			items, total, err = source(setter)
		}
	}
}
