package text

import (
	"database/sql/driver"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestText(t *testing.T) {
	var v Text
	v.Set("hello")
	value, err := v.Value()

	Then(
		t, "Text 提供 text 数据类型和值转换",
		Expect(Text("").DataType("sqlite"), Equal("text")),
		Expect(value, Equal(driver.Value("hello"))),
		Expect(err, Equal(error(nil))),
	)

	emptyValue, err := Text("").Value()
	Then(
		t, "空 Text 转换为 SQL NULL",
		Expect(emptyValue, Equal(driver.Value(nil))),
		Expect(err, Equal(error(nil))),
	)

	var fromString Text
	var fromBytes Text
	_ = fromString.Scan("from-string")
	_ = fromBytes.Scan([]byte("from-bytes"))

	Then(
		t, "Text 可扫描 string 与 []byte",
		Expect(fromString, Equal(Text("from-string"))),
		Expect(fromBytes, Equal(Text("from-bytes"))),
	)
}

func TestBlob(t *testing.T) {
	var v Blob
	v.Set([]byte("hello"))
	value, err := v.Value()

	Then(
		t, "Blob 根据驱动选择数据类型并转换非空值",
		Expect(Blob(nil).DataType("postgres"), Equal("bytea")),
		Expect(Blob(nil).DataType("sqlite"), Equal("blob")),
		Expect(value, Equal(driver.Value([]byte("hello")))),
		Expect(err, Equal(error(nil))),
	)

	emptyValue, err := Blob(nil).Value()
	Then(
		t, "空 Blob 转换为 SQL NULL",
		Expect(emptyValue, Equal(driver.Value(nil))),
		Expect(err, Equal(error(nil))),
	)

	var fromString Blob
	var fromBytes Blob
	_ = fromString.Scan("from-string")
	_ = fromBytes.Scan([]byte("from-bytes"))

	Then(
		t, "Blob 可扫描 string 与 []byte",
		Expect([]byte(fromString), Equal([]byte("from-string"))),
		Expect([]byte(fromBytes), Equal([]byte("from-bytes"))),
	)
}
