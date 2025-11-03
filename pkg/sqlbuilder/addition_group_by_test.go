package sqlbuilder_test

import (
	"testing"

	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
)

func TestGroupBy(t *testing.T) {
	tx := T("t_x")

	t.Run("select group by", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).From(
				tx,
				Where(TypedCol[int]("F_a").V(Eq(1))),
				GroupBy(Col("F_a")).Having(TypedCol[int]("F_a").V(Eq(1))),
			),

			testutil.BeFragment(`
SELECT *
FROM t_x
WHERE f_a = ?
GROUP BY f_a HAVING f_a = ?
`, 1, 1),
		)
	})

	t.Run("select desc group by", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					tx,
					Where(TypedCol[int]("F_a").V(Eq(1))),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
			testutil.BeFragment(`
SELECT *
FROM t_x
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
				1,
			))
	})
	t.Run("select multi group by", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					tx,
					Where(TypedCol[int]("F_a").V(Eq(1))),
					GroupBy(AscOrder(Col("F_a")), DescOrder(Col("F_b"))),
				),
			testutil.BeFragment(`
SELECT *
FROM t_x
WHERE f_a = ?
GROUP BY (f_a) ASC,(f_b) DESC
`,
				1,
			))
	})
}
