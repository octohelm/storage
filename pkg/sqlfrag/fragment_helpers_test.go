package sqlfrag_test

import (
	"context"
	"database/sql/driver"
	"iter"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	sqlfrag "github.com/octohelm/storage/pkg/sqlfrag"
)

type ctxKey struct{}

type injector struct{}

func (injector) InjectContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, "ok")
}

type customArg string

func (c customArg) ValueEx() string {
	return "ST_GeomFromText(?)"
}

type valuer string

func (v valuer) Value() (driver.Value, error) {
	return string(v), nil
}

func TestConstFuncAndContext(t *testing.T) {
	c := sqlfrag.Const("x")
	q, _ := sqlfrag.Collect(context.Background(), c)
	Then(t, "Const 和 Empty 暴露基本 fragment 行为",
		Expect(q, Equal("x")),
		Expect(c.IsNil(), Equal(false)),
		Expect(sqlfrag.Empty().IsNil(), Equal(true)),
	)

	fn := sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
		return func(yield func(string, []any) bool) {
			yield(ctx.Value(ctxKey{}).(string), nil)
		}
	})
	wrapped := sqlfrag.WithContextInjector(injector{}, fn)
	q, _ = sqlfrag.Collect(context.Background(), wrapped)

	Then(t, "Func 和 WithContextInjector 允许定制上下文输出",
		Expect(fn.IsNil(), Equal(false)),
		Expect(q, Equal("ok")),
	)
}

func TestBlockArgAndOperators(t *testing.T) {
	q, args := sqlfrag.Collect(context.Background(), sqlfrag.JoinValues(", ", sqlfrag.Const("a"), sqlfrag.Pair("?", 1)))
	Then(t, "JoinValues 连接常量和参数片段",
		Expect(q, Equal("a, ?")),
		Expect(args, Equal([]any{1})),
	)

	q, _ = sqlfrag.Collect(context.Background(), sqlfrag.Block(sqlfrag.Pair("SELECT 1")))
	Then(t, "Block 和 InlineBlock 保留括号语义",
		Expect(q, Equal("(SELECT 1\n)")),
	)

	argQ, argArgs := sqlfrag.Collect(context.Background(), sqlfrag.Pair("@v", sqlfrag.NamedArgSet{"v": 1}))
	Then(t, "NamedArgSet 通过命名占位符展开参数",
		Expect(argQ, Equal("?")),
		Expect(argArgs, Equal([]any{1})),
	)

	orderQ, _ := sqlfrag.Collect(sqlbuilder.ContextWithToggles(context.Background(), sqlbuilder.Toggles{
		sqlbuilder.ToggleUseValues: true,
	}), sqlfrag.Join("", sqlfrag.NonNil(iter.Seq[sqlfrag.Fragment](func(yield func(sqlfrag.Fragment) bool) {
		yield(sqlfrag.Const("a"))
		yield(nil)
	}))))
	Then(t, "NonNil 会跳过空 fragment",
		Expect(orderQ, Equal("a")),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", customArg("POINT(0 0)")))
	Then(t, "CustomValueArg 可替换占位片段",
		Expect(q, Equal("ST_GeomFromText(?)")),
		Expect(args, Equal([]any{customArg("POINT(0 0)")})),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", valuer("v")))
	Then(t, "driver.Valuer 保持单参数形式",
		Expect(q, Equal("?")),
		Expect(args, Equal([]any{valuer("v")})),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", []int{1, 2}))
	Then(t, "切片参数会按多个占位符展开",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{1, 2})),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", []byte("ab")))
	Then(t, "字节切片不会按序列拆开",
		Expect(q, Equal("?")),
		Expect(args, Equal([]any{[]byte("ab")})),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", iter.Seq[any](func(yield func(any) bool) {
		yield(1)
		yield(2)
	})))
	Then(t, "iter.Seq 参数可按值序列展开",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{1, 2})),
	)

	q, _ = sqlfrag.Collect(context.Background(), sqlfrag.InlineBlock(sqlfrag.Const("x")))
	Then(t, "InlineBlock 使用内联右括号",
		Expect(q, Equal("(x)")),
	)

	q, _ = sqlfrag.Collect(context.Background(), sqlfrag.BlockWithoutBrackets(sqlfrag.Map(iter.Seq[int](func(yield func(int) bool) {
		yield(1)
		yield(2)
	}), func(v int) sqlfrag.Fragment {
		return sqlfrag.Pair("?", v)
	})))
	Then(t, "Map 和 BlockWithoutBrackets 支持片段序列展开",
		Expect(q, Equal("?,?")),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", []string{"a", "b"}))
	Then(t, "字符串切片也会展开为多个参数",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{"a", "b"})),
	)

	q, args = sqlfrag.Collect(context.Background(), sqlfrag.Pair("?", []any{"a", 1}))
	Then(t, "[]any 参数直接透传展开",
		Expect(q, Equal("?,?")),
		Expect(args, Equal([]any{"a", 1})),
	)

	Then(t, "Join(nil) 与 Empty 都是空 fragment",
		Expect(sqlfrag.Join(",", nil).IsNil(), Equal(true)),
		ExpectMustValue(func() (string, error) {
			q, _ := sqlfrag.Collect(context.Background(), sqlfrag.Empty())
			return q, nil
		}, Equal("")),
	)
}
