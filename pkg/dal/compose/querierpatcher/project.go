package querierpatcher

import (
	"context"

	"github.com/octohelm/storage/pkg/dal"
	dalcomposetarget "github.com/octohelm/storage/pkg/dal/compose/target"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	slicesx "github.com/octohelm/x/slices"
)

func InSelectIfExists[T any, M sqlbuilder.Model](
	ctx context.Context,
	col sqlbuilder.TypedColumn[T],
	patchers ...Typed[M],
) sqlbuilder.ColumnValuer[T] {
	return dal.InSelect(col, ApplyToQuerier(dal.From(dalcomposetarget.Table[M](ctx), dal.WhereStmtNotEmpty()).Select(col), patchers...))
}

func DistinctSelect[M sqlbuilder.Model](projects ...sqlfrag.Fragment) Typed[M] {
	return &selectPatcher[M]{
		distinct: true,
		projects: projects,
	}
}

func Select[M sqlbuilder.Model](projects ...sqlfrag.Fragment) Typed[M] {
	return &selectPatcher[M]{
		projects: projects,
	}
}

func Returning[M sqlbuilder.Model](projects ...modelscoped.Column[M]) dal.MutationPatcher[M] {
	return &selectPatcher[M]{
		projects: slicesx.Map(projects, func(col modelscoped.Column[M]) sqlfrag.Fragment {
			return col
		}),
	}
}

type selectPatcher[M sqlbuilder.Model] struct {
	modelscoped.M[M]

	distinct bool
	projects []sqlfrag.Fragment
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
