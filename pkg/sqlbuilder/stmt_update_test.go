package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestStmtUpdate(t *testing.T) {
	table := T("T")

	t.Run("update", func(t *testing.T) {
		fa := TypedCol[int]("F_a")
		fb := TypedCol[int]("F_b")

		testutil.ShouldBeExpr(t,
			Update(table).
				Set(
					fa.By(Value(1)),
					fb.By(Value(2)),
				).
				Where(
					fa.V(Eq(1)),
					Comment("Comment"),
				),
			`
UPDATE T SET f_a = ?, f_b = ?
WHERE f_a = ?
/* Comment */`, 1, 2, 1)
	})
}
