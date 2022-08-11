package sqlbuilder

import (
	"context"
	"strconv"
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

func (l *limit) Ex(ctx context.Context) *Ex {
	e := ExactlyExpr("LIMIT ")
	e.WriteQuery(strconv.FormatInt(l.rowCount, 10))

	if l.offset > 0 {
		e.WriteQuery(" OFFSET ")
		e.WriteQuery(strconv.FormatInt(l.offset, 10))
	}

	return e.Ex(ctx)
}
