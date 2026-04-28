package filter

import (
	"bytes"
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
)

type composedValueRule struct {
	skip string
	Nick Filter[string] `name:"nick,omitempty"`
}

type composedPointerRule struct {
	Age *Filter[int]
}

func TestComposedInternals(t *testing.T) {
	c := Compose(1, composedValueRule{
		Nick: *Eq("alice"),
	}, composedPointerRule{
		Age: Gt(18),
	})

	Then(t, "Compose 会跳过非结构体并接受值类型 RuleExpr",
		Expect(c.IsZero(), Equal(false)),
		Expect(len(c.rules), Equal(2)),
	)

	fr := &fieldRuler{
		tpe:         reflect.TypeFor[composedPointerRule](),
		name:        "Age",
		ruleExprIdx: 0,
	}
	wrapper := fr.New()
	dec := directive.NewDecoder(bytes.NewBufferString(`gt(21)`))
	dec.RegisterDirectiveNewer(directive.DefaultDirectiveNewer, func() directive.Unmarshaler {
		return &Filter[int]{}
	})

	Then(t, "fieldRuler 和 ruleWrapper 可创建并驱动具体规则",
		Expect(fr.Name(), Equal("Age")),
		Expect(wrapper.Obj() != nil, Equal(true)),
		Expect(wrapper.Rule() != nil, Equal(true)),
		ExpectDo(func() error { return wrapper.UnmarshalDirective(dec) }),
		ExpectMustValue(func() (string, error) {
			raw, err := wrapper.Rule().MarshalText()
			return string(raw), err
		}, Equal(`gt(21)`)),
	)

	composed := &Composed{}
	composed.register("nick", fr)
	Then(t, "register 会初始化 fieldRulers map",
		Expect(composed.fieldRulers["nick"] != nil, Equal(true)),
	)
}

func TestComposedUnmarshalErrors(t *testing.T) {
	c := Compose(composedPointerRule{})

	Then(t, "空文本直接返回 nil",
		ExpectDo(func() error { return c.UnmarshalText(nil) }),
	)

	err := c.UnmarshalText([]byte(`where(eq(1))`))
	Then(t, "缺少字段名时返回非法指令错误",
		Expect(err != nil, Equal(true)),
		Expect(err.Error(), Equal("unsupported directive: eq")),
	)

	err = c.UnmarshalText([]byte(`where(1,eq(1))`))
	Then(t, "字段名 JSON 非字符串时直接返回解析错误",
		Expect(err != nil, Equal(true)),
	)

	err = c.UnmarshalText([]byte(`noop()`))
	Then(t, "未知顶层指令会被忽略且不写入规则",
		Expect(err, Equal(error(nil))),
		Expect(c.IsZero(), Equal(true)),
	)
}
