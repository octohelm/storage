package sqlbuilder_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestFunctions(t *testing.T) {
	countQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Count())
	avgQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Avg(sqlfrag.Const("f_age")))
	anyValueQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.AnyValue(sqlfrag.Const("f_name")))
	distinctQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Distinct(sqlfrag.Const("f_name")))
	minQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Min(sqlfrag.Const("f_age")))
	maxQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Max(sqlfrag.Const("f_age")))
	firstQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.First(sqlfrag.Const("f_name")))
	lastQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Last(sqlfrag.Const("f_name")))
	sumQ, _ := sqlfrag.Collect(context.Background(), sqlbuilder.Sum(sqlfrag.Const("f_age")))

	Then(
		t, "聚合函数 helper 会生成对应 SQL",
		Expect(countQ, Equal("COUNT(1)")),
		Expect(avgQ, Equal("AVG(f_age)")),
		Expect(anyValueQ, Equal("ANY_VALUE(f_name)")),
		Expect(distinctQ, Equal("DISTINCT(f_name)")),
		Expect(minQ, Equal("MIN(f_age)")),
		Expect(maxQ, Equal("MAX(f_age)")),
		Expect(firstQ, Equal("FIRST(f_name)")),
		Expect(lastQ, Equal("LAST(f_name)")),
		Expect(sumQ, Equal("SUM(f_age)")),
		Expect(sqlbuilder.Func(""), Equal((*sqlbuilder.Function)(nil))),
	)
}
