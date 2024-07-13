package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Where[M sqlbuilder.Model](w sqlbuilder.SqlExpr) Typed[M] {
	return &wherePatcher[M]{SqlExpr: w}
}

type wherePatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	sqlbuilder.SqlExpr
}

func (w *wherePatcher[M]) Apply(q dal.Querier) dal.Querier {
	if w.SqlExpr == nil || w.SqlExpr.IsNil() {
		return q
	}
	return q.WhereAnd(sqlbuilder.SqlExpr(w))
}
