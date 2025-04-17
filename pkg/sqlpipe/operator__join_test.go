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

type OrgUser struct {
	model.OrgUser

	User model.User
	Org  model.Org
}

func TestSourceFromWithJoin(t *testing.T) {
	t.Run("exec", func(t *testing.T) {
		orgUser := sqlpipe.FromAll[OrgUser]().Pipe(
			sqlpipe.JoinOnAs[OrgUser](
				model.OrgUserT.UserID,
				model.UserT.ID,
			),
			sqlpipe.JoinOnAs[OrgUser](
				model.OrgUserT.OrgID,
				model.OrgT.ID,
				sqlpipe.Where(model.OrgT.ID, sqlbuilder.Neq(model.OrgID(0))),
			),
		)

		testingx.Expect[sqlfrag.Fragment](t, orgUser, testutil.BeFragment(`
SELECT *
FROM t_org_user
JOIN t_user ON t_org_user.f_user_id = t_user.f_id
JOIN t_org ON (t_org_user.f_org_id = t_org.f_id) AND ((t_org.f_id <> ?) AND (t_org.f_deleted_at = ?))
`, model.OrgID(0), int64(0)))
	})

	t.Run("then where", func(t *testing.T) {
		filtered := sqlpipe.FromAll[OrgUser]().Pipe(
			sqlpipe.JoinOnAs[OrgUser](
				model.OrgUserT.OrgID,
				model.OrgT.ID,
			),
			sqlpipe.JoinOnAs[OrgUser](
				model.OrgUserT.UserID,
				model.UserT.ID,
			),
			sqlpipe.CastWhere[OrgUser](model.UserT.Name, sqlbuilder.Eq("x")),
			sqlpipe.CastOrWhere[OrgUser](model.OrgT.Name, sqlbuilder.Neq("x")),
		)

		testingx.Expect[sqlfrag.Fragment](t, filtered, testutil.BeFragment(`
SELECT *
FROM t_org_user
JOIN t_org ON t_org_user.f_org_id = t_org.f_id
JOIN t_user ON t_org_user.f_user_id = t_user.f_id
WHERE (t_user.f_name = ?) OR (t_org.f_name <> ?)
`, "x", "x"))
	})
}
