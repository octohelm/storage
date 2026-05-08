package flags

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestFlag(t *testing.T) {
	f := IncludesAll.With(ForReturning)
	Then(
		t, "Flag 支持 With、Without 和 Is",
		Expect(f.Is(IncludesAll), Equal(true)),
		Expect(f.Is(ForReturning), Equal(true)),
		Expect(f.Without(IncludesAll).Is(IncludesAll), Equal(false)),
	)
}
