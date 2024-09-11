package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"
)

func TestAssignment(t *testing.T) {
	t.Run("ColumnsAndValues", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.WithContextInjector(
				Toggles{
					ToggleUseValues: true,
				},
				ColumnsAndValues(Cols("a", "b"), 1, 2, 3, 4),
			),
			testutil.BeFragment(
				"(a,b) VALUES (?,?),(?,?)",
				1, 2, 3, 4,
			),
		)
	})
}
