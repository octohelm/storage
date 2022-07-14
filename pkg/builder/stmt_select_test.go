package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestSelect(t *testing.T) {
	table := T("T")

	t.Run("select with modifier", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil, "DISTINCT").
				From(
					table,
					Where(
						Col("F_a").Eq(1),
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
						Col("F_a").Eq(1),
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
						Col("F_a").Eq(1),
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
				Where(Col("F_a").Eq(1)),
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
