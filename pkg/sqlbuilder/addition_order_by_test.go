package sqlbuilder_test

import (
	"testing"

	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
)

func TestOrderBy(t *testing.T) {
	table := T("T")

	t.Run("select Order", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
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
			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
ORDER BY (f_a) ASC,(f_b) DESC
`,
				1,
			))
	})
}
