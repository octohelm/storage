package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"

	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestStmtUpdate(t *testing.T) {
	table := T("T")

	t.Run("update", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Update(table).
				Set(
					Col("F_a").ValueBy(1),
					Col("F_b").ValueBy(2),
				).
				Where(
					Col("F_a").Eq(1),
					Comment("Comment"),
				),
			`
UPDATE T SET f_a = ?, f_b = ?
WHERE f_a = ?
/* Comment */`, 1, 2, 1)
	})
}
