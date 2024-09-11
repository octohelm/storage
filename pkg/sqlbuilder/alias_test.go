package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestAlias(t *testing.T) {
	t.Run("alias", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](
			t,
			Alias(sqlfrag.Pair("f_id"), "id"),
			testutil.BeFragment(
				"f_id AS id",
			),
		)
	})
}
