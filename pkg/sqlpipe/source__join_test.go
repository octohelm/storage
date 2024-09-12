package sqlpipe_test

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"testing"

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
	orgUser := sqlpipe.FromAll[OrgUser](
		sqlpipe.JoinOn[OrgUser](model.OrgUserT.UserID, model.UserT.ID),
		sqlpipe.JoinOn[OrgUser](model.OrgUserT.OrgID, model.OrgT.ID),
	)

	t.Run("exec", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t, orgUser, testutil.BeFragment(`
SELECT *
FROM t_org_user
JOIN t_user ON t_org_user.f_user_id = t_user.f_id
JOIN t_org ON t_org_user.f_org_id = t_org.f_id
`))
	})

	t.Run("then where", func(t *testing.T) {
		filtered := orgUser.Pipe(
			sqlpipe.CastWhere[OrgUser](model.UserT.Name, sqlbuilder.Eq("x")),
			sqlpipe.CastOrWhere[OrgUser](model.OrgT.Name, sqlbuilder.Neq("x")),
		)

		testingx.Expect[sqlfrag.Fragment](t, filtered, testutil.BeFragment(`
SELECT *
FROM t_org_user
JOIN t_user ON t_org_user.f_user_id = t_user.f_id
JOIN t_org ON t_org_user.f_org_id = t_org.f_id
WHERE (t_user.f_name = ?) OR (t_org.f_name <> ?)
`, "x", "x"))
	})
}
