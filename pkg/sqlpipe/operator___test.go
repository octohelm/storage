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

func TestOperator(t *testing.T) {
	t.Run("from table user", func(t *testing.T) {
		src := sqlpipe.From[model.User]()

		t.Run("with projects", func(t *testing.T) {
			filteredSrc := src.Pipe(
				sqlpipe.Select[model.User](
					model.UserT.ID,
					model.UserT.Age,
				),
			)

			testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT f_id, f_age
FROM t_user
WHERE f_deleted_at = ?
`, int64(0)))
		})

		t.Run("with where", func(t *testing.T) {
			filteredSrc := src.Pipe(
				sqlpipe.Where(model.UserT.Name, sqlbuilder.Eq("x")),
			)

			testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE (f_name = ?) AND (f_deleted_at = ?)
`, "x", int64(0)))
		})
	})

	t.Run("from table user with all", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]()

		t.Run("with projects", func(t *testing.T) {
			filteredSrc := src.Pipe(
				sqlpipe.Select[model.User](
					model.UserT.ID,
					model.UserT.Age,
				),
			)

			testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT f_id, f_age
FROM t_user
`))
		})

		t.Run("with distinct on", func(t *testing.T) {
			filteredSrc := src.Pipe(
				sqlpipe.DescSort(model.UserT.UpdatedAt),
				sqlpipe.Select[model.User](
					model.UserT.ID,
					model.UserT.Age,
				),
				sqlpipe.DistinctOn(model.UserT.Age),
			)

			testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT f_id, f_age
FROM (
	SELECT DISTINCT ON ( f_age ) t_user.*
	FROM t_user
	ORDER BY (f_age),(f_updated_at) DESC
) AS t_user
ORDER BY (f_updated_at) DESC
`))
		})

		t.Run("then where", func(t *testing.T) {
			filteredSrc := src.Pipe(
				sqlpipe.Where(model.UserT.Name, sqlbuilder.Eq("x")),
			)

			t.Run("exec", func(t *testing.T) {
				testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_name = ?
`, "x"))
			})

			t.Run("then where compose", func(t *testing.T) {
				filteredSrc2 := filteredSrc.Pipe(
					sqlpipe.Where[model.User](model.UserT.ID, sqlbuilder.Eq[model.UserID](1)),
				)

				testingx.Expect[sqlfrag.Fragment](t, filteredSrc2, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE (f_name = ?) AND (f_id = ?)
`, "x", model.UserID(1)))
			})

			t.Run("then where sort", func(t *testing.T) {
				src2 := filteredSrc.Pipe(
					sqlpipe.AscSort[model.User](model.UserT.Name),
					sqlpipe.DescSort[model.User](model.UserT.ID),
				)

				testingx.Expect[sqlfrag.Fragment](t, src2, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_name = ?
ORDER BY (f_name) ASC,(f_id) DESC
`, "x"))
			})

			t.Run("then limit", func(t *testing.T) {
				limitedSrc := sqlpipe.Pipe(
					filteredSrc,
					sqlpipe.Limit[model.User](10),
				)

				testingx.Expect[sqlfrag.Fragment](t, limitedSrc, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_name = ?
LIMIT 10
`, "x"))
			})
		})

		t.Run("then limit", func(t *testing.T) {
			limitedSrc := src.Pipe(
				sqlpipe.Limit[model.User](10),
			)

			t.Run("exec", func(t *testing.T) {
				testingx.Expect[sqlfrag.Fragment](t, limitedSrc, testutil.BeFragment(`
SELECT *
FROM t_user
LIMIT 10
`))
			})

			t.Run("then where", func(t *testing.T) {
				filteredSrc := sqlpipe.Pipe(
					sqlpipe.As(limitedSrc, "t_user"),
					sqlpipe.Where[model.User](model.UserT.Name, sqlbuilder.Eq("x")),
				)

				testingx.Expect[sqlfrag.Fragment](t, filteredSrc, testutil.BeFragment(`
SELECT *
FROM (
	SELECT *
	FROM t_user
	LIMIT 10
) AS t_user
WHERE f_name = ?
`, "x"))
			})
		})
	})
}
