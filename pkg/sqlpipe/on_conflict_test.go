package sqlpipe

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
	"github.com/octohelm/storage/testdata/model"
)

func TestOnConflictBranches(t *testing.T) {
	base := &onConflictSource[model.User]{
		cols: model.UserT.I.IName,
	}

	q, _ := sqlfrag.Collect(context.Background(), base.toOnConflictAddition(context.Background(), 0))
	Then(t, "默认 on conflict 走 DO NOTHING",
		Expect(q, Equal("ON CONFLICT (f_name,f_deleted_at) DO NOTHING")),
	)

	update := &onConflictSource[model.User]{
		cols:    model.UserT.I.IName,
		updates: []modelscoped.Column[model.User]{model.UserT.Name},
	}
	q, _ = sqlfrag.Collect(context.Background(), update.toOnConflictAddition(context.Background(), 0))
	Then(t, "显式 updates 会生成 DO UPDATE SET",
		Expect(q, Equal("ON CONFLICT (f_name,f_deleted_at) DO UPDATE SET f_name = EXCLUDED.f_name")),
	)

	withFn := &onConflictSource[model.User]{
		cols: model.UserT.I.IName,
		with: func(on sqlbuilder.OnConflictAddition) sqlbuilder.OnConflictAddition {
			return on.DoUpdateSet(sqlbuilder.ColumnsAndValues(model.UserT.Name, "alice"))
		},
	}
	q, args := sqlfrag.Collect(context.Background(), withFn.toOnConflictAddition(context.Background(), 0))
	Then(t, "自定义 with 回调优先于默认策略",
		Expect(q, Equal("ON CONFLICT (f_name,f_deleted_at) DO UPDATE SET f_name = ?")),
		Expect(args, Equal([]any{"alice"})),
	)

	q, _ = sqlfrag.Collect(context.Background(), base.toOnConflictAddition(context.Background(), flags.ForReturning))
	Then(t, "返回行场景会回退为自更新冲突列",
		Expect(q, Equal("ON CONFLICT (f_name,f_deleted_at) DO UPDATE SET f_name = EXCLUDED.f_name")),
	)
}
