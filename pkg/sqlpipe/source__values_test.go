package sqlpipe_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
	testingx "github.com/octohelm/x/testing"
)

func TestSourceValues(t *testing.T) {
	t.Run("noop", func(t *testing.T) {
		src := sqlpipe.Values(make([]*model.User, 0), model.UserT.Name).Pipe(
			sqlpipe.OnConflictDoNothing(model.UserT.I.IName),
		)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(``))
	})

	t.Run("insert", func(t *testing.T) {
		users := []*model.User{
			{
				Name: "1",
			},
			{
				Name: "2",
			},
		}

		src := sqlpipe.Values(users, model.UserT.Name)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
INSERT INTO t_user (f_name)
VALUES
	(?),
	(?)
`,
			"1", "2",
		))

		t.Run("then on conflict", func(t *testing.T) {
			withOnConflict := src.Pipe(
				sqlpipe.OnConflictDoNothing(model.UserT.I.IName),
			)

			testingx.Expect[sqlfrag.Fragment](t, withOnConflict, testutil.BeFragment(`
INSERT INTO t_user (f_name)
VALUES
	(?),
	(?)
ON CONFLICT (f_name,f_deleted_at) DO NOTHING
`,
				"1", "2",
			))

			t.Run("then returning", func(t *testing.T) {
				withReturning := withOnConflict.Pipe(
					sqlpipe.Returning[model.User](),
				)

				testingx.Expect[sqlfrag.Fragment](t, withReturning, testutil.BeFragment(`
INSERT INTO t_user (f_name)
VALUES
	(?),
	(?)
ON CONFLICT (f_name,f_deleted_at) DO NOTHING
RETURNING *
`,
					"1", "2",
				))
			})
		})

		t.Run("then returning", func(t *testing.T) {
			withReturning := src.Pipe(
				sqlpipe.Returning[model.User](),
			)

			testingx.Expect[sqlfrag.Fragment](t, withReturning, testutil.BeFragment(`
INSERT INTO t_user (f_name)
VALUES
	(?),
	(?)
RETURNING *
`,
				"1", "2",
			))
		})
	})

	t.Run("insert omit", func(t *testing.T) {
		users := []*model.User{
			{
				Name: "1",
			},
			{
				Name: "2",
			},
		}

		src := sqlpipe.ValuesOmit(users,
			model.UserT.Nickname,
			model.UserT.Username,
			model.UserT.Gender,
			model.UserT.Age,
			model.UserT.CreatedAt,
			model.UserT.UpdatedAt,
			model.UserT.DeletedAt,
		)
		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
INSERT INTO t_user (f_name)
VALUES
	(?),
	(?)
`,
			"1", "2",
		))
	})
}
