package testutil

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestFragmentMatcher(t *testing.T) {
	m := BeFragment("SELECT 1", 1).(*fragmentMatcher[sqlfrag.Fragment])
	m2 := BeFragmentForQuery("SELECT 2", 1).(*fragmentMatcher[sqlfrag.Fragment])

	Then(t, "matcher 暴露固定动作和非负向语义",
		Expect(m.Action(), Equal("Be Frag")),
		Expect(m.Negative(), Equal(false)),
	)

	Then(t, "nil fragment 仅匹配空 query",
		Expect(m.Match(sqlfrag.Empty()), Equal(false)),
		Expect(BeFragment("").Match(sqlfrag.Empty()), Equal(true)),
	)

	Then(t, "query 不匹配时会记录规范化输出",
		Expect(m.Match(sqlfrag.Const("SELECT 9")), Equal(false)),
		Expect(m.NormalizeActual(sqlfrag.Const("SELECT 9")), Equal(any("SELECT 9"))),
		Expect(m.NormalizedExpected(), Equal(any("SELECT 1"))),
	)

	Then(t, "BeFragmentForQuery 在 query 不匹配时也会展示参数",
		Expect(m2.Match(sqlfrag.Pair("SELECT ?", 1)), Equal(false)),
		Expect(m2.NormalizeActual(sqlfrag.Pair("SELECT ?", 1)), Equal(any("SELECT ? | [1]"))),
		Expect(m2.NormalizedExpected(), Equal(any("SELECT 2 | [1]"))),
	)
}
