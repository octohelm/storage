package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
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
