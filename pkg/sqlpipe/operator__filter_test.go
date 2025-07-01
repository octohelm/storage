package sqlpipe_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
	modelfilter "github.com/octohelm/storage/testdata/model/filter"
	testingx "github.com/octohelm/x/testing"
)

func TestWhereInSelectFrom(t *testing.T) {
	t.Run("should build in select", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.WhereInSelectFrom(
				model.UserT.ID, model.OrgUserT.UserID,
				sqlpipe.From[model.OrgUser]().Pipe(
					sqlpipe.WhereInSelectFrom(
						model.OrgUserT.OrgID, model.OrgT.ID,
						sqlpipe.From[model.Org]().Pipe(
							&modelfilter.OrgByName{
								Name: filter.Eq("name"),
							},
						),
					),
				),
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_id IN (
	SELECT f_user_id
	FROM t_org_user
	WHERE f_org_id IN (
		SELECT f_id
		FROM t_org
		WHERE (f_name = ?) AND (f_deleted_at = ?)
	)
)
`, "name", int64(0)))
	})

	t.Run("should where not soft deleted", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.WhereInSelectFrom(
				model.UserT.ID, model.OrgUserT.UserID,
				sqlpipe.From[model.OrgUser]().Pipe(
					sqlpipe.WhereInSelectFrom(
						model.OrgUserT.OrgID, model.OrgT.ID,
						sqlpipe.From[model.Org](),
					),
				),
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_id IN (
	SELECT f_user_id
	FROM t_org_user
	WHERE f_org_id IN (
		SELECT f_id
		FROM t_org
		WHERE f_deleted_at = ?
	)
)
`, int64(0)))
	})

	t.Run("should build when limit exists", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.WhereInSelectFrom(
				model.UserT.ID,
				model.OrgUserT.UserID,
				sqlpipe.From[model.OrgUser]().Pipe(
					sqlpipe.Limit[model.OrgUser](10),
				),
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
SELECT *
FROM t_user
WHERE f_id IN (
	SELECT f_user_id
	FROM t_org_user
	LIMIT 10
)
`))
	})

	t.Run("should not build when where empty", func(t *testing.T) {
		src := sqlpipe.FromAll[model.User]().Pipe(
			sqlpipe.WhereInSelectFrom(
				model.UserT.ID,
				model.OrgUserT.UserID,
				sqlpipe.From[model.OrgUser]().Pipe(
					&modelfilter.OrgUserByOrgID{},
				),
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, src, testutil.BeFragment(`
SELECT *
FROM t_user
	`))
	})
}
