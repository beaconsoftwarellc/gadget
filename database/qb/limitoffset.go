package qb

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

// LimitOffset provides named limit and offset accessors
type LimitOffset interface {
	// Limit the number of records returned by a query
	Limit() uint
	// Offset the number of records to skip before returning results
	Offset() uint
}

// LimitOffsetSetter provides named limit and offset
// accessors and setters for use with queries.
type LimitOffsetSetter[T constraints.Integer] interface {
	LimitOffset
	// SetLimit to the passed value
	SetLimit(limit T) LimitOffsetSetter[T]
	// SetOffset to the passed value
	SetOffset(offset T) LimitOffsetSetter[T]
}

// NewLimitOffset with the specified Integer type
func NewLimitOffset[T constraints.Integer]() LimitOffsetSetter[T] {
	return &limitOffset[T]{
		offset: defaultOffset,
		limit:  defaultLimit,
	}
}

type limitOffset[T constraints.Integer] struct {
	limit  uint
	offset uint
}

func (lo *limitOffset[T]) SetLimit(limit T) LimitOffsetSetter[T] {
	if limit < 0 {
		// use an error so we get a stack trace
		_ = log.Warn(errors.Newf(
			"invalid (limit=%d)<0 and will be ignored", limit))
		return lo
	}
	lo.limit = intutil.ClampCast[T, uint](limit)
	return lo
}

func (lo *limitOffset[T]) SetOffset(offset T) LimitOffsetSetter[T] {
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
