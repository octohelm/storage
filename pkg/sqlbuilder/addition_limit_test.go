package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestLimit(t *testing.T) {
	table := T("T")

	t.Run("select limit", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1),
				), `
SELECT * FROM T
WHERE f_a = ?
LIMIT 1
`, 1)
	})
	t.Run("select without limit", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(-1),
				), `
SELECT * FROM T
WHERE f_a = ?
`, 1,
		)
	})
	t.Run("select limit and offset", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(
						Col("F_a").Eq(1),
					),
					Limit(1).Offset(200),
				),
			`
SELECT * FROM T
WHERE f_a = ?
LIMIT 1 OFFSET 200
`,
			1,
		)
	})
}
