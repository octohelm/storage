package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestStmtDelete(t *testing.T) {
	table := T("T")

	t.Run("delete", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Delete().From(table,
				Where(TypedCol[int]("F_a").V(Eq(1))),
				Comment("Comment"),
			),
			testutil.BeFragment(`
DELETE FROM T
WHERE f_a = ?
/* Comment */
`, 1),
		)
	})
}
