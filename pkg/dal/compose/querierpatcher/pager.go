package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Offset[M sqlbuilder.Model](offset int64) Typed[M] {
	return &offsetPatcher[M]{offset: offset}
}

type offsetPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	offset int64
}

func (w *offsetPatcher[M]) Apply(q dal.Querier) dal.Querier {
	if w.offset == 0 {
		return q
	}
	return q.Offset(w.offset)
}

func Limit[M sqlbuilder.Model](limit int64) Typed[M] {
	return &limitPatcher[M]{limit: limit}
}

type limitPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	limit int64
}

func (w *limitPatcher[M]) Apply(q dal.Querier) dal.Querier {
	if w.limit < 0 {
		return q
	}
	return q.Limit(w.limit)
}
