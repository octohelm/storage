package builder_test

import (
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestAlias(t *testing.T) {
	t.Run("alias", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Alias(Expr("f_id"), "id"), "f_id AS id")
	})
}
