package internal

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type StmtPatcher[M sqlbuilder.Model] interface {
	ApplyStmt(context.Context, *Builder[M]) *Builder[M]
}

type StmtPatcherFunc[M sqlbuilder.Model] func(context.Context, *Builder[M]) *Builder[M]

func (fn StmtPatcherFunc[M]) ApplyStmt(ctx context.Context, b *Builder[M]) *Builder[M] {
	return fn(ctx, b)
}
