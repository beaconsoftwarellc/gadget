package sliceutil

import (
	"iter"

	"github.com/beaconsoftwarellc/gadget/v2/database/qb"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
)

const maxIterations = 10000

// Flatten is a utility function for flattening a loop that ranges over
// an array of an array of items returned from a function. Useful in paging
// queries.
func Flatten[T any](limitOffset qb.LimitOffset, source func(limitOffset qb.LimitOffset) ([]T, int, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var (
			items  []T
			total  int
			err    error
			setter = qb.NewLimitOffset[uint]().SetOffset(limitOffset.Offset()).SetLimit(limitOffset.Limit())
			i      int
		)
		for i = 0; i < maxIterations; i++ {
			var (
				item T
			)
			items, total, err = source(setter)
			if err != nil {
				if !yield(item, err) {
					return
				}
				err = nil
			}
			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}
			// we retry the page on err + continue
			setter.SetOffset(setter.Offset() + uint(len(items)))
			if total < 0 {
				// negative total means we're done
				// this would wrap to a large number on the uint cast
				total = 0
			}
			// only check the offset against total if the error was not
			// skipped by the caller. We treat all values other than the
			// error as indeterminate.
			if setter.Offset() >= uint(total) && err == nil {
				break
			}
		}
		if i >= maxIterations {
			var zero T
			yield(zero, errors.New("flattening loop exceeded max iterations"))
		}
	}
}
