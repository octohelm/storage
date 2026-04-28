package bool

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestBool(t *testing.T) {
	var b Bool
	Then(t, "Bool 支持布尔转换和 schema 类型",
		Expect(FromBool(true), Equal(BOOL_TRUE)),
		Expect(FromBool(false), Equal(BOOL_FALSE)),
		Expect(BOOL_TRUE.Bool(), Equal(true)),
		Expect(BOOL_UNKNOWN.Bool(), Equal(false)),
		Expect(BOOL_TRUE.OpenAPISchemaType(), Equal([]string{"boolean"})),
	)

	Then(t, "Bool 支持文本和 JSON 编解码",
		ExpectMustValue(func() (string, error) {
			raw, err := BOOL_FALSE.MarshalJSON()
			return string(raw), err
		}, Equal("false")),
		ExpectDo(func() error { return b.UnmarshalText([]byte(`"true"`)) }),
	)
	Then(t, "带引号文本解析后可继续序列化",
		Expect(b, Equal(BOOL_TRUE)),
		ExpectDo(func() error { return b.UnmarshalJSON([]byte("false")) }),
		ExpectMustValue(func() (string, error) {
			raw, err := b.MarshalText()
			return string(raw), err
		}, Equal("false")),
	)
}
