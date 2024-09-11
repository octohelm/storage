package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestFunc(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t, Func(""), testutil.BeFragment(""))
	})

	t.Run("count", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t, Count(), testutil.BeFragment("COUNT(1)"))
	})
	t.Run("AVG", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t, Avg(), testutil.BeFragment("AVG(*)"))
	})
}
