package filter_test

import (
	"context"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	filter "github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestLitAndHelpers(t *testing.T) {
	v := filter.Lit("alice")
	Then(t, "Lit 暴露值和指针值",
		Expect(v.Value(), Equal("alice")),
		ExpectMustValue(func() (string, error) {
			raw, err := v.(interface{ MarshalJSON() ([]byte, error) }).MarshalJSON()
			return string(raw), err
		}, Equal(`"alice"`)),
		Expect(*v.(interface{ PtrValue() *string }).PtrValue(), Equal("alice")),
	)

	var decoded filter.Filter[int]
	Then(t, "Filter 可从字面值文本解析",
		ExpectDo(func() error { return decoded.UnmarshalText([]byte(`1`)) }),
	)
	Then(t, "字面值解析后保留 Eq 参数",
		Expect(decoded.Op(), Equal(filter.OP__EQ)),
		Expect(len(slices.Collect(decoded.Args())), Equal(1)),
	)

	var quoted filter.Filter[string]
	Then(t, "带引号文本会走字符串字面值解析",
		ExpectDo(func() error { return quoted.UnmarshalText([]byte(`"alice"`)) }),
	)
	Then(t, "字符串字面值解析后仍保留 Eq 规则",
		Expect(quoted.String(), Equal(`eq("alice")`)),
		ExpectMustValue(func() (string, error) {
			raw, err := slices.Collect(quoted.Args())[0].MarshalText()
			return string(raw), err
		}, Equal(`"alice"`)),
	)
}

func TestFilterFragmentsFromValues(t *testing.T) {
	f := filter.Eq(1)
	arg := slices.Collect(f.Args())[0].(filter.Value[int]).Value()
	frag := sqlfrag.Pair("?", arg)
	q, args := sqlfrag.Collect(context.Background(), frag)
	Then(t, "Values 可直接参与 SQL fragment 组装",
		Expect(q, Equal("?")),
		Expect(args, Equal([]any{1})),
	)
}
