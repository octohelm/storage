package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Where[M sqlbuilder.Model](w sqlbuilder.SqlExpr) interface {
	TypedQuerierPatcher[M]
	dal.MutationPatcher[M]
} {
	return &wherePatcher[M]{SqlExpr: w}
}

type wherePatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	sqlbuilder.SqlExpr
}

func (w *wherePatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	if w.SqlExpr == nil || w.SqlExpr.IsNil() {
		return q
	}
	return q.WhereAnd(sqlbuilder.SqlExpr(w))
}

func (w *wherePatcher[M]) ApplyMutation(m dal.Mutation[M]) dal.Mutation[M] {
	if w.SqlExpr == nil || w.SqlExpr.IsNil() {
		return m
	}
	return m.WhereAnd(sqlbuilder.SqlExpr(w))
}
