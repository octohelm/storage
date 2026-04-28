package json

import (
	"database/sql/driver"
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

type payload struct {
	Name string `json:"name,omitzero"`
	Age  int    `json:"age,omitzero"`
}

func TestValue(t *testing.T) {
	data := &payload{Name: "alice", Age: 18}
	v := ValueOf(data)

	dbValue, err := v.Value()
	Then(t, "Value 包装指针并序列化为数据库文本",
		Expect(v.IsZero(), Equal(false)),
		Expect(v.Get(), Equal(data)),
		Expect(v.OneOf()[0], Equal(any(&payload{}))),
		Expect(dbValue, Equal(driver.Value(`{"name":"alice","age":18}`))),
		Expect(err, Equal(error(nil))),
		Expect(v.DataType("sqlite"), Equal("text")),
	)

	var copied payload
	v.As(&copied)
	Then(t, "As 将非空值复制到目标",
		Expect(copied, Equal(*data)),
	)

	var decoded Value[payload]
	Then(t, "Value 可从 JSON 恢复",
		ExpectDo(func() error {
			return decoded.UnmarshalJSON([]byte(`{"name":"bob"}`))
		}),
	)
	Then(t, "Value JSON 反序列化后保留指针和值",
		Expect(decoded.Get() != nil, Equal(true)),
		ExpectMustValue(func() (payload, error) {
			return *decoded.Get(), nil
		}, Equal(payload{Name: "bob"})),
	)

	Then(t, "Value 可从数据库值恢复",
		ExpectDo(func() error {
			return decoded.Scan([]byte(`{"name":"carol","age":20}`))
		}),
	)
	Then(t, "Value Scan 后保留指针和值",
		Expect(decoded.Get() != nil, Equal(true)),
		ExpectMustValue(func() (payload, error) {
			return *decoded.Get(), nil
		}, Equal(payload{Name: "carol", Age: 20})),
	)

	decoded.Set(&payload{Name: "dave"})
	raw, err := decoded.MarshalJSON()
	Then(t, "MarshalJSON 使用当前指针值",
		Expect(string(raw), Equal(`{"name":"dave"}`)),
		Expect(err, Equal(error(nil))),
	)

	nullValue, err := Value[payload]{}.Value()
	Then(t, "空 Value 写入数据库时转换为空字符串",
		Expect(Value[payload]{}.IsZero(), Equal(true)),
		Expect(nullValue, Equal(driver.Value(""))),
		Expect(err, Equal(error(nil))),
	)
}

func TestObject(t *testing.T) {
	obj := ObjectOf(&payload{Name: "alice"})

	dbValue, err := obj.Value()
	Then(t, "Object 包装内联对象并提供数据库文本",
		Expect(obj.IsZero(), Equal(false)),
		Expect(obj.OneOf()[0], Equal(any(&payload{}))),
		Expect(dbValue, Equal(driver.Value(`{"name":"alice"}`))),
		Expect(err, Equal(error(nil))),
		Expect(obj.DataType("postgres"), Equal("text")),
	)

	var copied payload
	obj.As(&copied)
	Then(t, "Object.As 复制非空数据",
		Expect(copied, Equal(payload{Name: "alice"})),
	)

	var decoded Object[payload]
	Then(t, "Object 可从 JSON 恢复",
		ExpectDo(func() error {
			return decoded.UnmarshalJSON([]byte(`{"name":"bob"}`))
		}),
	)
	Then(t, "Object JSON 反序列化后保留数据",
		Expect(decoded.Data != nil, Equal(true)),
		ExpectMustValue(func() (payload, error) {
			return *decoded.Data, nil
		}, Equal(payload{Name: "bob"})),
	)

	Then(t, "Object 可从 string 数据库值扫描",
		ExpectDo(func() error {
			return decoded.Scan(`{"name":"carol","age":20}`)
		}),
	)
	Then(t, "Object Scan 后保留数据",
		Expect(decoded.Data != nil, Equal(true)),
		ExpectMustValue(func() (payload, error) {
			return *decoded.Data, nil
		}, Equal(payload{Name: "carol", Age: 20})),
	)

	Then(t, "Object 扫描 nil 时保持现有数据",
		ExpectDo(func() error {
			return decoded.Scan(nil)
		}),
		ExpectDo(func() error {
			return nil
		}),
	)

	decoded.Set(&payload{Name: "dave"})
	raw, err := decoded.MarshalJSON()
	Then(t, "Object.MarshalJSON 使用当前数据",
		Expect(string(raw), Equal(`{"name":"dave"}`)),
		Expect(err, Equal(error(nil))),
	)
}

func TestArray(t *testing.T) {
	values := Array[payload]{{Name: "alice"}, {Name: "bob", Age: 20}}

	dbValue, err := values.Value()
	Then(t, "Array 提供 JSON 文本数据库值",
		Expect(values.IsZero(), Equal(false)),
		Expect(values.DataType("sqlite"), Equal("text")),
		Expect(dbValue, Equal(driver.Value(`[{"name":"alice"},{"name":"bob","age":20}]`))),
		Expect(err, Equal(error(nil))),
	)

	var decoded Array[payload]
	Then(t, "Array 可扫描数据库 JSON 文本",
		ExpectDo(func() error {
			return decoded.Scan([]byte(`[{"name":"carol"}]`))
		}),
	)
	Then(t, "Array Scan 后保留元素",
		Expect(decoded, Equal(Array[payload]{{Name: "carol"}})),
	)

	emptyValue, err := Array[payload]{}.Value()
	Then(t, "空 Array 写入数据库时转换为空字符串",
		Expect(Array[payload]{}.IsZero(), Equal(true)),
		Expect(emptyValue, Equal(driver.Value(""))),
		Expect(err, Equal(error(nil))),
	)
}

func TestScanValueErrors(t *testing.T) {
	var data payload

	Then(t, "scanValue 拒绝非字符串数据库值",
		ExpectDo(func() error {
			return scanValue(1, &data)
		}, ErrorMatch(regexp.MustCompile("cannot sql\\.Scan\\(\\)"))),
	)

	Then(t, "空数据库文本不覆盖现有值",
		ExpectDo(func() error {
			return scanValue("", &data)
		}),
		Expect(data, Equal(payload{})),
	)
}
