package filter

import (
	"context"
	"iter"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	rootfilter "github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
)

func TestBuildWhere(t *testing.T) {
	f := rootfilter.And(
		rootfilter.Eq("alice"),
		rootfilter.Prefix("a"),
	)

	frag := BuildWhere(f, func(op rootfilter.Op, seq iter.Seq[string], create func(iter.Seq[string]) sqlbuilder.ColumnValuer[string]) sqlfrag.Fragment {
		return model.UserT.Name.V(create(seq))
	})

	q, args := sqlfrag.Collect(context.Background(), frag)
	Then(
		t, "BuildWhere 组合 AND 和字符串操作",
		Expect(q, Equal("(f_name = ?) AND (f_name LIKE ?)")),
		Expect(args, Equal([]any{"alice", "a%"})),
	)

	inFrag := BuildWhere(rootfilter.In("a", "b"), func(op rootfilter.Op, seq iter.Seq[string], create func(iter.Seq[string]) sqlbuilder.ColumnValuer[string]) sqlfrag.Fragment {
		return model.UserT.Name.V(create(seq))
	})
	q, args = sqlfrag.Collect(context.Background(), inFrag)
	Then(
		t, "BuildWhere 支持 IN",
		Expect(q, Equal("f_name IN (?,?)")),
		Expect(args, Equal([]any{"a", "b"})),
	)
}

func TestAsWhereAndHelpers(t *testing.T) {
	src := sqlpipe.FromAll[model.User]().Pipe(
		AsWhere(model.UserT.Name, rootfilter.Contains("ali")),
	)

	q, args := sqlfrag.Collect(context.Background(), src)
	Then(
		t, "AsWhere 把 Filter 转换为 SourceOperator",
		Expect(q, Equal("SELECT *\nFROM t_user\nWHERE f_name LIKE ?")),
		Expect(args, Equal([]any{"%ali%"})),
	)

	sub := slices.Collect(SubFilters(rootfilter.Or(rootfilter.Eq("a"), rootfilter.Eq("b"))))
	values := slices.Collect(Values(rootfilter.In("a", "b")))

	Then(
		t, "SubFilters 和 Values 提取子规则和值",
		Expect(len(sub), Equal(2)),
		Expect(values, Equal([]string{"a", "b"})),
	)

	Then(
		t, "空 Filter 不生成 where 条件",
		Expect(BuildWhere[string](nil, nil), Equal(sqlfrag.Fragment(nil))),
	)
}
