package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestGroupBy(t *testing.T) {
	table := T("T")

	t.Run("select group by", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(TypedCol[int]("F_a").V(Eq(1))),
					GroupBy(Col("F_a")).
						Having(TypedCol[int]("F_a").V(Eq(1))),
				),
			`SELECT * FROM T
WHERE f_a = ?
GROUP BY f_a HAVING f_a = ?
`,
			1, 1,
		)
	})

	t.Run("select desc group by", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(TypedCol[int]("F_a").V(Eq(1))),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
			`
SELECT * FROM T
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
			1,
		)
	})
	t.Run("select multi group by", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					Where(TypedCol[int]("F_a").V(Eq(1))),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),

			`
SELECT * FROM T
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
			1,
		)
	})
}
