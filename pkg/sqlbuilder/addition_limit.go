package sqlbuilder

import (
	"context"
	"fmt"
	"iter"
)

type LimitAddition interface {
	Addition

	Offset(offset int64) LimitAddition
}

func Limit(rowCount int64) LimitAddition {
	return &limit{rowCount: rowCount}
}

type limit struct {
	// LIMIT
	rowCount int64
	// OFFSET
	offset int64
}

func (l *limit) AdditionType() AdditionType {
	return AdditionLimit
}

func (l limit) Offset(offset int64) LimitAddition {
	l.offset = offset
	return &l
}

func (l *limit) IsNil() bool {
	return l == nil || l.rowCount <= 0
}

func (l *limit) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield(fmt.Sprintf("LIMIT %d", l.rowCount), nil) {
			return
		}

		if l.offset > 0 {
			if !yield(fmt.Sprintf(" OFFSET %d", l.offset), nil) {
				return
			}
		}
	}
}
