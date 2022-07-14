package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestStmtDelete(t *testing.T) {
	table := T("T")

	t.Run("delete", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Delete().From(table,
				Where(Col("F_a").Eq(1)),
				Comment("Comment"),
			),
			`
DELETE FROM T
WHERE f_a = ?
/* Comment */
`, 1)
	})
}
