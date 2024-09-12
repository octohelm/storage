package sqlpipe_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
	testingx "github.com/octohelm/x/testing"
)

func TestInsertFrom(t *testing.T) {
	src := sqlpipe.FromAll[model.User]().Pipe(
		sqlpipe.Where(model.UserT.Name, sqlbuilder.Eq("x")),
	)

	t.Run("do insert from", func(t *testing.T) {
		i := sqlpipe.InsertFrom(
			src.Pipe(
				sqlpipe.Project[model.User](
					model.UserT.Name,
					sqlfrag.Pair("?", 1),
				)),
			model.OrgT.Name,
			model.OrgT.ID,
		).Pipe(
			sqlpipe.OnConflictDoNothing(model.OrgT.I.IName),
		)

		testingx.Expect[sqlfrag.Fragment](t, i, testutil.BeFragment(`
INSERT INTO t_org (f_name,f_id) 
SELECT f_name, ?
FROM t_user
WHERE f_name = ?
ON CONFLICT (f_name) DO NOTHING 
`, 1, "x"))
	})
}
