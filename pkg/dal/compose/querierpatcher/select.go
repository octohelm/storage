package querierpatcher

import (
	"context"
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func InSelectIfExists[T any, M sqlbuilder.Model](
	ctx context.Context,
	col sqlbuilder.TypedColumn[T],
	patchers ...Typed[M],
) sqlbuilder.ColumnValueExpr[T] {
	t := col.T()
	s := dal.SessionFor(ctx, t)

	return dal.InSelect(col, ApplyTo(dal.From(s.T(t), dal.WhereStmtNotEmpty()).Select(col), patchers...))
}

func DistinctSelect[M sqlbuilder.Model](projects ...sqlbuilder.SqlExpr) Typed[M] {
	return &selectPatcher[M]{
		distinct: true,
		projects: projects,
	}
}

func Select[M sqlbuilder.Model](projects ...sqlbuilder.SqlExpr) Typed[M] {
	return &selectPatcher[M]{projects: projects}
}

type selectPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	distinct bool
	projects []sqlbuilder.SqlExpr
}

func (w *selectPatcher[M]) Apply(q dal.Querier) dal.Querier {
	if w.projects == nil {
		return q
	}

	if w.distinct {
		q = q.Distinct()
	}

	return q.Select(w.projects...)
}
