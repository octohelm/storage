package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestOrderBy(t *testing.T) {
	table := T("T")

	t.Run("select Order", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Select(nil).
				From(
					table,
					OrderBy(
						AscOrder(Col("F_a")),
						DescOrder(Col("F_b")),
					),
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
				),
			`
SELECT * FROM T
WHERE f_a = ?
ORDER BY (f_a) ASC,(f_b) DESC
`,
			1,
		)
	})
}
