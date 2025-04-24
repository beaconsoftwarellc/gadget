package database

import (
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/intutil"
	"github.com/beaconsoftwarellc/gadget/v2/log"
	"golang.org/x/exp/constraints"
)

const (
	defaultLimit  uint = 50
	defaultOffset uint = 0
)

// LimitOffset provides named limit and offset
// accessors and setters for use with queries.
type LimitOffset[T constraints.Integer] interface {
	SetLimit(limit T) LimitOffset[T]
	SetOffset(offset T) LimitOffset[T]
	Limit() uint
	Offset() uint
}

// NewLimitOffset with the specified Integer type
func NewLimitOffset[T constraints.Integer]() LimitOffset[T] {
	return &limitOffset[T]{
		offset: defaultOffset,
		limit:  defaultLimit,
	}
}

type limitOffset[T constraints.Integer] struct {
	limit  uint
	offset uint
}

func (lo *limitOffset[T]) SetLimit(limit T) LimitOffset[T] {
	if limit < 0 {
		// use an error so we get a stack trace
		_ = log.Warn(errors.Newf(
			"invalid (limit=%d)<0 and will be ignored", limit))
		return lo
	}
	lo.limit = intutil.ClampCast[T, uint](limit)
	return lo
}

func (lo *limitOffset[T]) SetOffset(offset T) LimitOffset[T] {
	if offset < 0 {
		// use an error so we get a stack trace
		_ = log.Warn(errors.Newf(
			"invalid (offset=%d)<0 and will be ignored", offset))
		return lo
	}
	lo.offset = intutil.ClampCast[T, uint](offset)
	return lo
}

func (lo *limitOffset[T]) Limit() uint {
	return lo.limit
}

func (lo *limitOffset[T]) Offset() uint {
	return lo.offset
}
