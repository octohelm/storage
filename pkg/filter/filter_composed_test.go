package filter_test

import (
	"iter"
	"net/http"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	filter "github.com/octohelm/storage/pkg/filter"
)

type byNick struct {
	Nick *filter.Filter[string] `name:"nick"`
}

type byAge struct {
	Age *filter.Filter[int]
}

func TestComposeAndUtil(t *testing.T) {
	c := filter.Compose(byNick{}, byAge{})

	Then(t, "Compose 先注册字段定义但零值不生成规则",
		Expect(c.IsZero(), Equal(true)),
	)

	Then(t, "Compose 可从 where/or 文本恢复结构化过滤器",
		ExpectDo(func() error {
			return c.UnmarshalText([]byte(`or(where("nick",eq("alice")),where("Age",gt(18)))`))
		}),
	)

	Then(t, "Compose 反序列化后保留规则和过滤对象",
		Expect(c.IsZero(), Equal(false)),
		Expect(len(c.Filters), Equal(2)),
		ExpectMustValue(func() (string, error) {
			raw, err := c.MarshalText()
			return string(raw), err
		}, Equal(`or(where("nick",eq("alice")),where("Age",gt(18)))`)),
	)

	items := []filter.Arg{
		filter.Eq(1),
		filter.Where("age", filter.Gt(2)),
		filter.Gte(3),
	}

	Then(t, "MapFilter、MapWhere 与 First 只返回匹配项",
		Expect(slices.Collect(filter.MapFilter[int](slices.Values(items), func(f *filter.Filter[int]) (string, bool) {
			return f.String(), true
		})), Equal([]string{"eq(1)", "gte(3)"})),
		Expect(slices.Collect(filter.MapWhere(slices.Values(items), func(arg filter.Arg) (string, bool) {
			if rule, ok := arg.(filter.Rule); ok && rule.Op() == filter.OP__WHERE {
				raw, _ := rule.MarshalText()
				return string(raw), true
			}
			return "", false
		})), Equal([]string{`where("age",gt(2))`})),
		ExpectMustValue(func() (string, error) {
			v, ok := filter.First(slices.Values(items), func(arg filter.Arg) (string, bool) {
				if rule, ok := arg.(filter.Rule); ok && rule.Op() == filter.OP__GTE {
					raw, _ := rule.MarshalText()
					return string(raw), true
				}
				return "", false
			})
			if !ok {
				return "", http.ErrMissingFile
			}
			return v, nil
		}, Equal("gte(3)")),
	)

	unsupported := filter.Compose(byNick{})
	err := unsupported.UnmarshalText([]byte(`or(where("missing",eq("alice")))`))
	Then(t, "Compose 遇到未注册字段时返回明确错误",
		Expect(err != nil, Equal(true)),
		Expect(err.Error(), Equal("unsupported ql field `missing`")),
	)

	preloaded := filter.Compose(byNick{Nick: filter.Eq("alice")})
	Then(t, "Compose 会把非零初始过滤器转成单条 where",
		Expect(preloaded.IsZero(), Equal(false)),
		ExpectMustValue(func() (string, error) {
			raw, err := preloaded.MarshalText()
			return string(raw), err
		}, Equal(`where("nick",eq("alice"))`)),
	)
}

func TestFilterConstructorsAndText(t *testing.T) {
	Then(t, "Filter 构造器写入正确 op",
		Expect(filter.Eq(1).Op(), Equal(filter.OP__EQ)),
		Expect(filter.Neq(1).Op(), Equal(filter.OP__NEQ)),
		Expect(filter.Lt(1).Op(), Equal(filter.OP__LT)),
		Expect(filter.Lte(1).Op(), Equal(filter.OP__LTE)),
		Expect(filter.Gt(1).Op(), Equal(filter.OP__GT)),
		Expect(filter.Gte(1).Op(), Equal(filter.OP__GTE)),
		Expect(filter.Contains("x").Op(), Equal(filter.OP__CONTAINS)),
		Expect(filter.Prefix("x").Op(), Equal(filter.OP__PREFIX)),
		Expect(filter.Suffix("x").Op(), Equal(filter.OP__SUFFIX)),
		Expect(filter.In(1, 2).Op(), Equal(filter.OP__IN)),
		Expect(filter.Notin(1, 2).Op(), Equal(filter.OP__NOTIN)),
		Expect(filter.And(filter.Eq(1), filter.Eq(2)).Op(), Equal(filter.OP__AND)),
		Expect(filter.Or(filter.Eq(1), filter.Eq(2)).Op(), Equal(filter.OP__OR)),
		Expect(filter.Intersection(filter.Eq(1), filter.Eq(2)).Op(), Equal(filter.OP__INTERSECTION)),
		Expect(filter.OrRules(filter.Eq(1), filter.Eq(2)).Op(), Equal(filter.OP__OR)),
	)

	seq := filter.InSeq(iter.Seq[int](func(yield func(int) bool) {
		yield(1)
		yield(2)
	}))
	notSeq := filter.NotinSeq(iter.Seq[int](func(yield func(int) bool) {
		yield(3)
		yield(4)
	}))
	Then(t, "InSeq 和 NotinSeq 收集序列值",
		Expect(len(slices.Collect(seq.Args())), Equal(2)),
		Expect(len(slices.Collect(notSeq.Args())), Equal(2)),
	)

	f := filter.Filter[int]{}
	Then(t, "Filter 支持文本和字面值解析",
		ExpectDo(func() error { return f.UnmarshalText([]byte("eq(1)")) }),
	)
	Then(t, "函数解析后保留 op 和参数",
		Expect(f.Op(), Equal(filter.OP__EQ)),
		Expect(len(slices.Collect(f.Args())), Equal(1)),
		Expect(f.String(), Equal("eq(1)")),
		Expect(f.New(), Equal(new(int))),
		Expect(len(f.OneOf()), Equal(1)),
	)

	Then(t, "纯字面值按 Eq 处理",
		ExpectDo(func() error { return f.UnmarshalText([]byte(`2`)) }),
	)
	Then(t, "字面值解析后转成 Eq",
		Expect(f.Op(), Equal(filter.OP__EQ)),
		Expect(len(slices.Collect(f.Args())), Equal(1)),
	)
}

func TestWhereAndArgAccessors(t *testing.T) {
	w := filter.Where("age", filter.Eq(1), filter.Gt(2))
	text, _ := w.MarshalText()

	Then(t, "Where 暴露参数、操作符和文本",
		Expect(w.Op(), Equal(filter.OP__WHERE)),
		Expect(w.IsZero(), Equal(false)),
		Expect(len(slices.Collect(w.Args())), Equal(2)),
		Expect(string(text), Equal(`where("age",eq(1),gt(2))`)),
		Expect(w.New(), Equal(new(int))),
	)

	var parsed filter.Rule = filter.Where[int]("")
	Then(t, "Where 可从文本反序列化",
		Expect(parsed != nil, Equal(true)),
	)
}
