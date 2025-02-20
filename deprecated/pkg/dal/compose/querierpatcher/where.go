package querierpatcher

import (
	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Where[M sqlbuilder.Model](w sqlfrag.Fragment) interface {
	Typed[M]
	dal.MutationPatcher[M]
} {
	return &wherePatcher[M]{Fragment: w}
}

type wherePatcher[M sqlbuilder.Model] struct {
	modelscoped.M[M]

	sqlfrag.Fragment
}

func (w *wherePatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	if sqlfrag.IsNil(w.Fragment) {
		return q
	}
	return q.WhereAnd(sqlfrag.Fragment(w))
}

func (w *wherePatcher[M]) ApplyMutation(m dal.Mutation[M]) dal.Mutation[M] {
	if sqlfrag.IsNil(w.Fragment) {
		return m
	}
	return m.WhereAnd(sqlfrag.Fragment(w))
}
