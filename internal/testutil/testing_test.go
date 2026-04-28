package testutil_test

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
	. "github.com/octohelm/x/testing/v2"

	testutil "github.com/octohelm/storage/internal/testutil"
)

func TestHelpers(t *testing.T) {
	ctx := testutil.NewContext(t)
	Then(t, "NewContext 返回可用上下文",
		Expect(ctx != nil, Equal(true)),
	)

	testutil.Expect(t, 1, testingx.Equal(1))
	testutil.Expect(t, 1, testutil.Not(testingx.Equal(2)))
	testutil.Expect(t, "x", testutil.Be("x"))
	testutil.Expect(t, []int{1, 2}, testutil.HaveLen[[]int](2))
}
