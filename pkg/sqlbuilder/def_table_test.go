package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestTable_Expr(t *testing.T) {
	tUser := T("t_user",
		Col("f_id", ColField("ID"), ColTypeOf(uint64(0), ",autoincrement")),
		Col("f_name", ColField("Name"), ColTypeOf("", ",size=128,default=''")),
	)

	tUserRole := T("t_user_role",
		Col("f_id", ColField("ID"), ColTypeOf(uint64(0), ",autoincrement")),
		Col("f_user_id", ColField("UserID"), ColTypeOf(uint64(0), "")),
	)

	t.Run("replace table", func(t *testing.T) {
		testingx.Expect(t,
			tUser.(TableCanFragment).Fragment("#.*"),
			testutil.BeFragment("t_user.*"),
		)
	})

	t.Run("replace table col by field", func(t *testing.T) {
		testingx.Expect(t, tUser.(TableCanFragment).Expr("#ID = #ID + 1"),
			testutil.BeFragment("f_id = f_id + 1"))
	})

	t.Run("replace table col by field for function", func(t *testing.T) {
		testingx.Expect(t,
			tUser.(TableCanFragment).Fragment("COUNT(#ID)"),
			testutil.BeFragment("COUNT(f_id)"))
	})

	t.Run("could handle context", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					tUser,
					Where(
						AsCond(tUser.(TableCanFragment).Fragment("#ID > 1")),
					),
					Join(tUserRole).On(AsCond(tUser.(TableCanFragment).Fragment("#ID = ?", tUserRole.(TableCanFragment).Fragment("#UserID")))),
				),
			testutil.BeFragment(`
SELECT *
FROM t_user
JOIN t_user_role ON t_user.f_id = t_user_role.f_user_id
WHERE t_user.f_id > 1
`))
	})
}
