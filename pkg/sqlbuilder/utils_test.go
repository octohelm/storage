package sqlbuilder

import (
	testingx "github.com/octohelm/x/testing"
	"testing"
)

func TestParseIndexDefine(t *testing.T) {
	i := ParseIndexDefine("index i_xxx,GIST TEST,gist_trgm_ops")

	testingx.Expect(t, i, testingx.Equal(&IndexDefine{
		Kind:   "index",
		Name:   "i_xxx",
		Method: "GIST",
		FieldNameAndOptions: []FieldNameAndOption{
			"TEST,gist_trgm_ops",
		},
	}))

	testingx.Expect(t, i.FieldNameAndOptions[0].Name(), testingx.Be("TEST"))
	testingx.Expect(t, i.FieldNameAndOptions[0].Options(), testingx.Equal([]string{"GIST_TRGM_OPS"}))
}
