package sqlbuilder_test

import (
	"slices"
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"

	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestStmtInsert(t *testing.T) {
	table := T("t_x", Col("f_a"), Col("f_b"))

	t.Run("insert with modifier", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Insert("IGNORE").
				Into(table).
				Values(Cols("f_a", "f_b"), 1, 2),

			testutil.BeFragment(`
INSERT IGNORE INTO t_x (f_a,f_b)
VALUES
	(?,?)
`, 1, 2))
	})

	t.Run("insert simple", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Insert().
				Into(table, Comment("Comment")).
				Values(Cols("f_a", "f_b"), 1, 2),
			testutil.BeFragment(`
INSERT INTO t_x (f_a,f_b)
VALUES
	(?,?)
/* Comment */
`, 1, 2))
	})

	t.Run("multiple insert", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Insert().
				Into(table).
				Values(Cols("f_a", "f_b"), 1, 2, 1, 2, 1, 2),
			testutil.BeFragment(`
INSERT INTO t_x (f_a,f_b)
VALUES
	(?,?),
	(?,?),
	(?,?)
`, 1, 2, 1, 2, 1, 2))
	})

	t.Run("multiple insert by iter", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Insert().
				Into(table).
				ValuesCollect(Cols("f_a", "f_b"), slices.Values([]any{1, 2, 1, 2, 1, 2})),
			testutil.BeFragment(`
INSERT INTO t_x (f_a,f_b)
VALUES
	(?,?),
	(?,?),
	(?,?)
`, 1, 2, 1, 2, 1, 2))
	})

	t.Run("insert from select", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Insert().
				Into(table).
				Values(Cols("f_a", "f_b"),
					Select(Cols("f_a", "f_b")).
						From(table, Where(TypedColOf[int](table, "f_a").V(Eq(1))))),
			testutil.BeFragment(`
INSERT INTO t_x (f_a,f_b) 
SELECT f_a,f_b
FROM t_x
WHERE f_a = ?
`, 1))
	})
}
