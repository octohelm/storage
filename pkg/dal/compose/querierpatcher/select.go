package querierpatcher

import (
	"context"
	"github.com/octohelm/storage/pkg/dal"
	dalcomposetarget "github.com/octohelm/storage/pkg/dal/compose/target"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func InSelectIfExists[T any, M sqlbuilder.Model](
	ctx context.Context,
	col sqlbuilder.TypedColumn[T],
	patchers ...Typed[M],
) sqlbuilder.ColumnValueExpr[T] {
	return dal.InSelect(col, ApplyTo(dal.From(dalcomposetarget.Table[M](ctx), dal.WhereStmtNotEmpty()).Select(col), patchers...))
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
