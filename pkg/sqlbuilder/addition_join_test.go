package sqlbuilder_test

import (
	"testing"

	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
)

func TestJoin(t *testing.T) {
	tUser := T("t_user",
		Col("f_id", ColTypeOf(uint64(0), ",autoincrement")),
		Col("f_name", ColTypeOf("", ",size=128,default=''")),
		Col("f_org_id", ColTypeOf("", ",size=128,default=''")),
	)

	tOrg := T("t_org",
		Col("f_org_id", ColTypeOf(uint64(0), ",autoincrement")),
		Col("f_org_name", ColTypeOf("", ",size=128,default=''")),
	)

	t.Run("JOIN ON", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(sqlfrag.JoinValues(", ",
				Alias(tUser.F("f_id"), "f_id"),
				Alias(tUser.F("f_name"), "f_name"),
				Alias(tUser.F("f_org_id"), "f_org_id"),
				Alias(tOrg.F("f_org_name"), "f_org_name"),
			)).From(
				tUser,
				Join(Alias(tOrg, "t_org")).On(
					TypedColOf[int](tUser, "f_org_id").V(
						EqCol(TypedColOf[int](tOrg, "f_org_id")),
					),
				),
			),
			testutil.BeFragment(`
SELECT t_user.f_id AS f_id, t_user.f_name AS f_name, t_user.f_org_id AS f_org_id, t_org.f_org_name AS f_org_name
FROM t_user
JOIN t_org AS t_org ON t_user.f_org_id = t_org.f_org_id
`,
			))
	})
	t.Run("JOIN USING", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					tUser,
					Join(tOrg).Using(tUser.F("f_org_id")),
				),
			testutil.BeFragment(`
SELECT *
FROM t_user
JOIN t_org USING (f_org_id)
`,
			))
	})
}
