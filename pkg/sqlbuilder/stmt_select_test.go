package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestSelect(t *testing.T) {
	table := T("T")

	t.Run("select with modifier", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil, Expr("DISTINCT")).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
				),
			`
SELECT DISTINCT * FROM T
WHERE f_a = ?`, 1)
	})
	t.Run("select simple", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
					Comment("comment"),
				),
			`
SELECT * FROM T
WHERE f_a = ?
/* comment */
`, 1,
		)
	})
	t.Run("select with target", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(Col("F_a")).
				From(table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
				),
			`
SELECT f_a FROM T
WHERE f_a = ?`, 1)
	})

	t.Run("select for update", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).From(
				table,
				Where(TypedCol[int]("F_a").V(Eq(1))),
				ForUpdate(),
			),
			`
SELECT * FROM T
WHERE f_a = ?
FOR UPDATE
`,
			1,
		)
	})
}
