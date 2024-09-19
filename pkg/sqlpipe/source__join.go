package sqlpipe

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func JoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
) FromOptionFunc[M] {
	return func(x *sourceFrom[M]) {
		x.patchers = append(x.patchers, internal.StmtPatcherFunc[M](func(ctx context.Context, b internal.StmtBuilder[M]) internal.StmtBuilder[M] {
			return b.WithAdditions(
				sqlbuilder.Join(b.T(ctx, new(S))).On(on.V(sqlbuilder.EqCol(src))))
		}))
	}
}
