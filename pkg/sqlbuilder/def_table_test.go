package sqlbuilder_test

import (
	"testing"

	//"github.com/octohelm/storage/internal/connectors/postgresql"
	"github.com/octohelm/storage/internal/testutil"
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
		testutil.ShouldBeExpr(t, tUser.(TableExprParse).Expr("#.*"), "t_user.*")
	})

	t.Run("replace table col by field", func(t *testing.T) {
		testutil.ShouldBeExpr(t, tUser.(TableExprParse).Expr("#ID = #ID + 1"), "f_id = f_id + 1")
	})

	t.Run("replace table col by field for function", func(t *testing.T) {
		testutil.ShouldBeExpr(t, tUser.(TableExprParse).Expr("COUNT(#ID)"), "COUNT(f_id)")
	})

	t.Run("could handle context", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					tUser,
					Where(
						AsCond(tUser.(TableExprParse).Expr("#ID > 1")),
					),
					Join(tUserRole).On(AsCond(tUser.(TableExprParse).Expr("#ID = ?", tUserRole.(TableExprParse).Expr("#UserID")))),
				),
			`
SELECT * FROM t_user
JOIN t_user_role ON t_user.f_id = t_user_role.f_user_id
WHERE t_user.f_id > 1
`)
	})
}
