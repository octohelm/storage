package patcher

import (
	"context"

	"github.com/octohelm/storage/pkg/dal"
	dalcomposetarget "github.com/octohelm/storage/pkg/dal/compose/target"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func InSelectIfExists[T any, M sqlbuilder.Model](
	ctx context.Context,
	col sqlbuilder.TypedColumn[T],
	patchers ...TypedQuerierPatcher[M],
) sqlbuilder.ColumnValueExpr[T] {
	return dal.InSelect(col, ApplyToQuerier(dal.From(dalcomposetarget.Table[M](ctx), dal.WhereStmtNotEmpty()).Select(col), patchers...))
}

func DistinctSelect[M sqlbuilder.Model](projects ...sqlbuilder.SqlExpr) TypedQuerierPatcher[M] {
	return &selectPatcher[M]{
		distinct: true,
		projects: projects,
	}
}

func Select[M sqlbuilder.Model](projects ...sqlbuilder.SqlExpr) TypedQuerierPatcher[M] {
	return &selectPatcher[M]{projects: projects}
}

func Returning[M sqlbuilder.Model](projects ...sqlbuilder.SqlExpr) dal.MutationPatcher[M] {
	return &selectPatcher[M]{projects: projects}
}

type selectPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	distinct bool
	projects []sqlbuilder.SqlExpr
}

func (w *selectPatcher[M]) ApplyMutation(m dal.Mutation[M]) dal.Mutation[M] {
	if w.projects == nil {
		return m
	}

	return m.Returning(w.projects...)
}

func (w *selectPatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	if w.projects == nil {
		return q
	}

	if w.distinct {
		q = q.Distinct()
	}

	return q.Select(w.projects...)
}
