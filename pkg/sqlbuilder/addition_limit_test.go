package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestLimit(t *testing.T) {
	table := T("T")

	t.Run("select limit", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
					Limit(1),
				),

			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
LIMIT 1
`, 1))
	})
	t.Run("select without limit", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
					Limit(-1),
				),
			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
`, 1,
			))
	})

	t.Run("select limit and offset", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Select(nil).
				From(
					table,
					Where(
						TypedCol[int]("F_a").V(Eq(1)),
					),
					Limit(10).Offset(200),
				),
			testutil.BeFragment(`
SELECT *
FROM T
WHERE f_a = ?
LIMIT 10 OFFSET 200
`,
				1,
			))
	})
}
