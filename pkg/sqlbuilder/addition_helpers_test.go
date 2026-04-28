package sqlbuilder_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestAdditionsAndFieldValueHelpers(t *testing.T) {
	ret := sqlbuilder.Returning(model.UserT.Name)
	onConflict := sqlbuilder.OnConflict(sqlbuilder.Cols("f_name")).DoNothing()
	onConflictUpdate := sqlbuilder.OnConflict(sqlbuilder.Cols("f_name")).DoUpdateSet(
		sqlbuilder.ColumnsAndValues(model.UserT.Name, "alice"),
	)
	orderBy := sqlbuilder.OrderBy(sqlbuilder.AscOrder(model.UserT.Name, sqlbuilder.NullsLast()), sqlbuilder.DescOrder(model.UserT.ID, sqlbuilder.NullsFirst()))
	distinct := sqlbuilder.DistinctOn(model.UserT.Name)

	retQ, _ := sqlfrag.Collect(context.Background(), ret)
	conflictQ, _ := sqlfrag.Collect(context.Background(), onConflict)
	conflictUpdateQ, conflictUpdateArgs := sqlfrag.Collect(context.Background(), onConflictUpdate)
	orderQ, _ := sqlfrag.Collect(context.Background(), orderBy)
	distinctQ, _ := sqlfrag.Collect(context.Background(), distinct)

	Then(t, "Returning、OnConflict、OrderBy、DistinctOn 生成片段",
		Expect(retQ, Equal("RETURNING f_name")),
		Expect(conflictQ, Equal("ON CONFLICT (f_name) DO NOTHING")),
		Expect(conflictUpdateQ, Equal("ON CONFLICT (f_name) DO UPDATE SET f_name = ?")),
		Expect(conflictUpdateArgs, Equal([]any{"alice"})),
		Expect(orderQ, Equal("ORDER BY (f_name) ASC NULLS LAST,(f_id) DESC NULLS FIRST")),
		Expect(distinctQ, Equal("DISTINCT ON (f_name)")),
		Expect(sqlbuilder.NullsFirst().IsNil(), Equal(false)),
		Expect(sqlbuilder.NullsLast().IsNil(), Equal(false)),
	)
}
