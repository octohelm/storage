package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestStmtUpdate(t *testing.T) {
	t0 := T("t_0",
		Col("f_a"),
		Col("f_b"),
	)
	t1 := T("t_1",
		Col("f_a"),
	)

	t.Run("update", func(t *testing.T) {
		fa := CastCol[int](t0.F("f_a"))
		fb := CastCol[int](t0.F("f_b"))

		testutil.ShouldBeExpr(t,
			Update(t0).
				Set(
					fa.By(Value(1)),
					fb.By(AsValue(CastCol[int](t1.F("f_a")))),
				).
				From(t1).
				Where(
					fa.V(Eq(1)),
					Comment("Comment"),
				),
			`
UPDATE t_0 SET f_a = ?, f_b = t_1.f_a FROM t_1
WHERE t_0.f_a = ?
/* Comment */`, 1, 1)
	})
}
