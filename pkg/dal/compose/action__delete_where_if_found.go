package compose

import (
	"context"
	"github.com/octohelm/storage/pkg/dal/compose/patcher"

	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func DeleteWhereIfFound[M sqlbuilder.Model, T any](ctx context.Context, idCol sqlbuilder.TypedColumn[T], patchers ...patcher.TypedQuerierPatcher[M]) error {
	return dal.Prepare(new(M)).ForDelete().
		Where(idCol.V(
			patcher.InSelectIfExists(ctx, idCol, patchers...),
		)).
		Save(ctx)
}
