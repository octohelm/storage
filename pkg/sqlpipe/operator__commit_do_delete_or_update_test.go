package sqlpipe_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/pkg/sqltype/time"
	"github.com/octohelm/storage/testdata/model"
	testingx "github.com/octohelm/x/testing"
)

func TestOperatorAction(t *testing.T) {
	src := sqlpipe.FromAll[model.User]().Pipe(
		sqlpipe.Where[model.User](model.UserT.Name, sqlbuilder.Eq("x")),
	)

	t.Run("do update with custom", func(t *testing.T) {
		updated := src.Pipe(
			sqlpipe.DoUpdate(model.UserT.Age, sqlbuilder.Incr[int64](1)),
			sqlpipe.DoUpdate(model.UserT.Name, sqlbuilder.Value("a")),
		)

		testingx.Expect[sqlfrag.Fragment](t, updated, testutil.BeFragment(`
UPDATE t_user
SET f_age = f_age + ?, f_name = ?
WHERE f_name = ?
`, int64(1), "a", "x"))
	})

	t.Run("do update", func(t *testing.T) {
		updated := src.Pipe(
			sqlpipe.DoUpdateSet(&model.User{
				Name: "a",
			},
				model.UserT.Name,
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, updated, testutil.BeFragment(`
UPDATE t_user
SET f_name = ?
WHERE f_name = ?
`, "a", "x"))
	})

	t.Run("do update", func(t *testing.T) {
		updated := src.Pipe(
			sqlpipe.DoUpdateSetOmitZero(&model.User{
				Name: "a",
			}),
		)

		testingx.Expect[sqlfrag.Fragment](t, updated, testutil.BeFragment(`
UPDATE t_user
SET f_name = ?
WHERE f_name = ?
`, "a", "x"))

		t.Run("then returning", func(t *testing.T) {
			withReturning := updated.Pipe(
				sqlpipe.Returning[model.User](),
			)

			testingx.Expect[sqlfrag.Fragment](t, withReturning, testutil.BeFragment(`
UPDATE t_user
SET f_name = ?
WHERE f_name = ?
RETURNING *
`, "a", "x"))
		})
	})

	t.Run("do delete", func(t *testing.T) {
		deleted := src.Pipe(
			sqlpipe.DoDeleteHard[model.User](),
		)

		testingx.Expect[sqlfrag.Fragment](t, deleted, testutil.BeFragment(`
DELETE FROM t_user
WHERE f_name = ?
`, "x"))

		t.Run("then returning", func(t *testing.T) {
			withReturning := deleted.Pipe(
				sqlpipe.Returning[model.User](),
			)

			testingx.Expect[sqlfrag.Fragment](t, withReturning, testutil.BeFragment(`
DELETE FROM t_user
WHERE f_name = ?
RETURNING *
`, "x",
			))
		})

		t.Run("soft", func(t *testing.T) {
			d := src.Pipe(
				sqlpipe.DoDelete[model.User](),
			)

			testingx.Expect[sqlfrag.Fragment](t, d, testutil.BeFragmentForQuery(`
UPDATE t_user
SET f_deleted_at = ?
WHERE (f_name = ?) AND (f_deleted_at = ?)
`, time.Now(), "x", 0))
		})
	})
}
