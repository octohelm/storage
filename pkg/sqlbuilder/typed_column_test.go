package sqlbuilder_test

import (
	"context"
	"iter"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestTypedColumnValuers(t *testing.T) {
	tUser := sqlbuilder.T("t_user",
		sqlbuilder.Col("f_id", sqlbuilder.ColField("ID")),
		sqlbuilder.Col("f_name", sqlbuilder.ColField("Name")),
	)

	idCol := sqlbuilder.TypedColOf[int](tUser, "f_id")
	nameCol := sqlbuilder.CastColumn[string](tUser.F("f_name"))
	created := sqlbuilder.TypedCol[string]("f_created", sqlbuilder.ColField("Created"))

	qEq, argsEq := sqlfrag.Collect(context.Background(), nameCol.V(sqlbuilder.Eq("alice")))
	qEqCol, _ := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.EqCol(sqlbuilder.TypedCol[int]("f_ref"))))
	qNeqCol, _ := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.NeqCol(sqlbuilder.TypedCol[int]("f_ref"))))
	qIn, argsIn := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.In(1, 2)))
	qInSeq, argsInSeq := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.InSeq(iter.Seq[int](func(yield func(int) bool) {
		yield(3)
		yield(4)
	}))))
	qNotIn, argsNotIn := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.NotIn(5, 6)))
	qNotInSeq, argsNotInSeq := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.NotInSeq(iter.Seq[int](func(yield func(int) bool) {
		yield(7)
		yield(8)
	}))))
	qNull, _ := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.IsNull[int]()))
	qNotNull, _ := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.IsNotNull[int]()))
	qLike, argsLike := sqlfrag.Collect(context.Background(), nameCol.V(sqlbuilder.Like("a")))
	qNotLike, argsNotLike := sqlfrag.Collect(context.Background(), nameCol.V(sqlbuilder.NotLike("a")))
	qLeftLike, argsLeftLike := sqlfrag.Collect(context.Background(), nameCol.V(sqlbuilder.LeftLike("a")))
	qRightLike, argsRightLike := sqlfrag.Collect(context.Background(), nameCol.V(sqlbuilder.RightLike("a")))
	qBetween, argsBetween := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.Between(1, 9)))
	qNotBetween, argsNotBetween := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.NotBetween(1, 9)))
	qGt, argsGt := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.Gt(1)))
	qGte, argsGte := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.Gte(1)))
	qLt, argsLt := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.Lt(9)))
	qLte, argsLte := sqlfrag.Collect(context.Background(), idCol.V(sqlbuilder.Lte(9)))

	Then(t, "TypedColumn 支持各类比较和匹配 valuer",
		Expect(qEq, Equal("f_name = ?")),
		Expect(argsEq, Equal([]any{"alice"})),
		Expect(qEqCol, Equal("f_id = f_ref")),
		Expect(qNeqCol, Equal("f_id <> f_ref")),
		Expect(qIn, Equal("f_id IN (?,?)")),
		Expect(argsIn, Equal([]any{1, 2})),
		Expect(qInSeq, Equal("f_id IN (?,?)")),
		Expect(argsInSeq, Equal([]any{3, 4})),
		Expect(qNotIn, Equal("f_id NOT IN (?,?)")),
		Expect(argsNotIn, Equal([]any{5, 6})),
		Expect(qNotInSeq, Equal("f_id NOT IN (?,?)")),
		Expect(argsNotInSeq, Equal([]any{7, 8})),
		Expect(qNull, Equal("f_id IS NULL")),
		Expect(qNotNull, Equal("f_id IS NOT NULL")),
		Expect(qLike, Equal("f_name LIKE ?")),
		Expect(argsLike, Equal([]any{"%a%"})),
		Expect(qNotLike, Equal("f_name NOT LIKE ?")),
		Expect(argsNotLike, Equal([]any{"%a%"})),
		Expect(qLeftLike, Equal("f_name LIKE ?")),
		Expect(argsLeftLike, Equal([]any{"%a"})),
		Expect(qRightLike, Equal("f_name LIKE ?")),
		Expect(argsRightLike, Equal([]any{"a%"})),
		Expect(qBetween, Equal("f_id BETWEEN ? AND ?")),
		Expect(argsBetween, Equal([]any{1, 9})),
		Expect(qNotBetween, Equal("f_id NOT BETWEEN ? AND ?")),
		Expect(argsNotBetween, Equal([]any{1, 9})),
		Expect(qGt, Equal("f_id > ?")),
		Expect(argsGt, Equal([]any{1})),
		Expect(qGte, Equal("f_id >= ?")),
		Expect(argsGte, Equal([]any{1})),
		Expect(qLt, Equal("f_id < ?")),
		Expect(argsLt, Equal([]any{9})),
		Expect(qLte, Equal("f_id <= ?")),
		Expect(argsLte, Equal([]any{9})),
	)

	qValue, argsValue := sqlfrag.Collect(context.Background(), created.By(sqlbuilder.Value("x")))
	qAsValue, _ := sqlfrag.Collect(context.Background(), created.By(sqlbuilder.AsValue(nameCol)))
	qIncr, argsIncr := sqlfrag.Collect(context.Background(), idCol.By(sqlbuilder.Incr(1)))
	qDes, argsDes := sqlfrag.Collect(context.Background(), idCol.By(sqlbuilder.Des(1)))

	Then(t, "Value、AsValue、Incr、Des 可直接生成 assignment",
		Expect(qValue, Equal("f_created = ?")),
		Expect(argsValue, Equal([]any{"x"})),
		Expect(qAsValue, Equal("f_created = f_name")),
		Expect(qIncr, Equal("f_id = f_id + ?")),
		Expect(argsIncr, Equal([]any{1})),
		Expect(qDes, Equal("f_id = f_id - ?")),
		Expect(argsDes, Equal([]any{1})),
	)

	Then(t, "空 In/NotIn 返回 nil fragment",
		Expect(sqlfrag.IsNil(sqlbuilder.In[int]()(idCol)), Equal(true)),
		Expect(sqlfrag.IsNil(sqlbuilder.NotIn[int]()(idCol)), Equal(true)),
	)
}
