package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestStmtInsert(t *testing.T) {
	table := T("T", Col("f_a"), Col("f_b"))

	t.Run("insert with modifier", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Insert("IGNORE").
				Into(table).
				Values(Cols("f_a", "f_b"), 1, 2),

			"INSERT IGNORE INTO T (f_a,f_b) VALUES (?,?)", 1, 2)
	})

	t.Run("insert simple", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Insert().
				Into(table, Comment("Comment")).
				Values(Cols("f_a", "f_b"), 1, 2),
			`
INSERT INTO T (f_a,f_b) VALUES (?,?)
/* Comment */
`, 1, 2)
	})

	t.Run("multiple insert", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Insert().
				Into(table).
				Values(Cols("f_a", "f_b"), 1, 2, 1, 2, 1, 2),
			"INSERT INTO T (f_a,f_b) VALUES (?,?),(?,?),(?,?)", 1, 2, 1, 2, 1, 2)
	})

	t.Run("insert from select", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Insert().
				Into(table).
				Values(Cols("f_a", "f_b"), Select(Cols("f_a", "f_b")).From(table, Where(table.F("f_a").Eq(1)))),
			`
INSERT INTO T (f_a,f_b) SELECT f_a,f_b FROM T
WHERE f_a = ?
`, 1)
	})
}
