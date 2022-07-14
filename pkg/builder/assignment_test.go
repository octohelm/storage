package builder_test

import (
	"context"
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"
	. "github.com/octohelm/sqlx/pkg/builder"
)

func TestAssignment(t *testing.T) {
	t.Run("ColumnsAndValues", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			ColumnsAndValues(Cols("a", "b"), 1, 2, 3, 4).Ex(ContextWithToggles(context.Background(), Toggles{
				ToggleUseValues: true,
			})),
			"(a,b) VALUES (?,?),(?,?)", 1, 2, 3, 4,
		)
	})
}
