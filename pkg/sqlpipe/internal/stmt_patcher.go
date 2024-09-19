package internal

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type StmtPatcher[M sqlbuilder.Model] interface {
	ApplyStmt(context.Context, StmtBuilder[M]) StmtBuilder[M]
}

type StmtPatcherFunc[M sqlbuilder.Model] func(context.Context, StmtBuilder[M]) StmtBuilder[M]

func (fn StmtPatcherFunc[M]) ApplyStmt(ctx context.Context, b StmtBuilder[M]) StmtBuilder[M] {
	return fn(ctx, b)
}
