package internal

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
	"github.com/octohelm/storage/testdata/model"
)

func TestHelpers(t *testing.T) {
	Then(
		t, "Flag 支持位运算",
		Expect(flags.IncludesAll.With(flags.ForReturning).Is(flags.IncludesAll), Equal(true)),
		Expect(flags.IncludesAll.Without(flags.IncludesAll).Is(flags.IncludesAll), Equal(false)),
	)

	ctx := FlagContext.Inject(context.Background(), flags.IncludesAll)
	Then(
		t, "Seed 从上下文合并 Flag",
		Expect(Seed{Flag: flags.ForReturning}.GetFlag(ctx), Equal(flags.IncludesAll|flags.ForReturning)),
		Expect(Seed{Flag: flags.ForReturning}.GetFlag(context.Background()), Equal(flags.ForReturning)),
	)

	patcher := StmtPatcherFunc[model.User](func(ctx context.Context, b *Builder[model.User]) *Builder[model.User] {
		return b.WithFlag(flags.ForReturning)
	})

	Then(
		t, "StmtPatcherFunc 实现 ApplyStmt",
		Expect(patcher.ApplyStmt(context.Background(), &Builder[model.User]{}).Flag, Equal(flags.ForReturning)),
	)

	additions := fixAdditions([]sqlbuilder.Addition{
		sqlbuilder.Limit(10),
		sqlbuilder.Returning(sqlfrag.Const("*")),
	})
	Then(
		t, "Returning 存在时移除 Limit",
		Expect(len(additions), Equal(1)),
		Expect(additions[0].AdditionType(), Equal(sqlbuilder.AdditionReturning)),
	)
}
