package sqlbuilder_test

import (
	"context"
	"iter"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestConditionsAndCombination(t *testing.T) {
	condQ, condArgs := sqlfrag.Collect(context.Background(), sqlbuilder.And(
		sqlbuilder.AsCond(sqlfrag.Pair("a = ?", 1)),
		sqlbuilder.Or(
			sqlfrag.Pair("b = ?", 2),
			sqlfrag.Pair("c = ?", 3),
		),
	))
	seqQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.XorSeq(iter.Seq[sqlfrag.Fragment](func(yield func(sqlfrag.Fragment) bool) {
		yield(sqlfrag.Pair("a = ?", 1))
		yield(sqlfrag.Pair("b = ?", 2))
	})))
	orQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.OrSeq(iter.Seq[sqlfrag.Fragment](func(yield func(sqlfrag.Fragment) bool) {
		yield(sqlfrag.Pair("a = ?", 1))
	})))

	unionQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Union().All(
		sqlbuilder.Select(sqlfrag.Const("1")),
	))
	intersectQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Intersect().Distinct(
		sqlbuilder.Select(sqlfrag.Const("2")),
	))
	exceptQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Expect().All(
		sqlbuilder.Select(sqlfrag.Const("3")),
	))

	Then(t, "条件组合和集合操作都会生成 SQL",
		Expect(condQ, Equal("(a = ?) AND ((b = ?) OR (c = ?))")),
		Expect(condArgs, Equal([]any{1, 2, 3})),
		Expect(seqQ, Equal("(a = ?) XOR (b = ?)")),
		Expect(orQ, Equal("a = ?")),
		Expect(strings.Contains(unionQ, "UNION ALL"), Equal(true)),
		Expect(strings.Contains(intersectQ, "INTERSECT DISTINCT"), Equal(true)),
		Expect(strings.Contains(exceptQ, "EXCEPT ALL"), Equal(true)),
		Expect(sqlbuilder.EmptyCond().IsNil(), Equal(true)),
	)
}
