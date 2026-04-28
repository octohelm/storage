package internal

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

// StmtPatcher 定义对 Builder 的补丁能力。
type StmtPatcher[M sqlbuilder.Model] interface {
	ApplyStmt(context.Context, *Builder[M]) *Builder[M]
}

// StmtPatcherFunc 允许直接用函数实现 StmtPatcher。
type StmtPatcherFunc[M sqlbuilder.Model] func(context.Context, *Builder[M]) *Builder[M]

func (fn StmtPatcherFunc[M]) ApplyStmt(ctx context.Context, b *Builder[M]) *Builder[M] {
	return fn(ctx, b)
}
