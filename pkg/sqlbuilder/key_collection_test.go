package sqlbuilder

import (
	"slices"
	"testing"

	testingx "github.com/octohelm/x/testing/v2"
)

func TestKeysInternal(t *testing.T) {
	testingx.Then(t, "空 keys 集合长度为零",
		testingx.Expect((*keys)(nil).Len(), testingx.Equal(0)),
	)

	ks := &keys{}
	ks.AddKey(nil, Index("i_name", Cols("f_name")))
	collected := slices.Collect(ks.Keys())

	testingx.Then(t, "AddKey 会跳过 nil，K 支持大小写无关匹配",
		testingx.Expect(ks.Len(), testingx.Equal(1)),
		testingx.Expect(ks.K("I_NAME").Name(), testingx.Equal("i_name")),
		testingx.Expect(ks.K("missing"), testingx.Equal(Key(nil))),
		testingx.Expect(len(collected), testingx.Equal(1)),
	)

	tUser := T("t_user", Col("f_name"))
	cloned := ks.Of(tUser)
	testingx.Then(t, "Of 会把 key 绑定到新表",
		testingx.Expect(cloned.Len(), testingx.Equal(1)),
		testingx.Expect(GetKeyTable(cloned.K("i_name")).TableName(), testingx.Equal("t_user")),
	)
}
