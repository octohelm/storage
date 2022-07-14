package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"

	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestFunc(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Func(""), "")
	})
	t.Run("count", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Count(), "COUNT(1)")
	})
	t.Run("AVG", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Avg(), "AVG(*)")
	})
}
