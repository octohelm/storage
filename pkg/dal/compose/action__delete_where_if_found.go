package compose

import (
	"context"

	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/dal/compose/querierpatcher"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func DeleteWhereIfFound[M sqlbuilder.Model, T any](ctx context.Context, idCol sqlbuilder.TypedColumn[T], patchers ...querierpatcher.Typed[M]) error {
	return dal.Prepare(new(M)).ForDelete().
		Where(idCol.V(
			querierpatcher.InSelectIfExists(ctx, idCol, patchers...),
		)).
		Save(ctx)
}
