package sqlbuilder_test

import (
	"testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"
)

func TestSelect(t *testing.T) {
	table := T("T")

	t.Run("select with modifier", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil, sqlfrag.Pair("DISTINCT")).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
				),
			testutil.BeFragment(`
SELECT DISTINCT *
FROM T
WHERE f_a = ?`, 1))
	})
	t.Run("select simple", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
					Comment("comment"),
				),
			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
/* comment */
`, 1,
			))
	})
	t.Run("select with target", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(Col("F_a")).
				From(table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
				),
			testutil.BeFragment(`
SELECT f_a
FROM T
WHERE f_a = ?`, 1))
	})

	t.Run("select for update", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).From(
				table,
				Where(TypedCol[int]("F_a").V(Eq(1))),
				ForUpdate(),
			),
			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
FOR UPDATE
`,
				1,
			))
	})
}
