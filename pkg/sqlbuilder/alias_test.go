package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestAlias(t *testing.T) {
	t.Run("alias", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Alias(Expr("f_id"), "id"), "f_id AS id")
	})
}
