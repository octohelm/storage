package time

import (
	"testing"
	"time"

	. "github.com/octohelm/x/testing/v2"
)

func TestTimestampHelpers(t *testing.T) {
	prevOutputLayout := outputLayout
	prevCST := CST
	prevLayouts := supportedLayouts
	t.Cleanup(func() {
		outputLayout = prevOutputLayout
		CST = prevCST
		supportedLayouts = prevLayouts
	})

	origin := Timestamp(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))
	SetOutputLayout(time.RFC3339, time.UTC)
	AddSupportedLayout("2006-01-02", time.UTC)

	Then(t, "Timestamp 支持加减与日期偏移",
		Expect(Add(origin, time.Hour).Unix(), Equal(origin.Unix()+3600)),
		Expect(Sub(Add(origin, time.Hour), origin), Equal(time.Hour)),
		Expect(AddDate(origin, 0, 0, 1).Day(), Equal(3)),
	)

	Then(t, "Timestamp 支持字符串解析和文本编解码",
		ExpectDo(func() error {
			_, err := ParseTimestampFromStringWithLayout("2024-01-02", "2006-01-02")
			return err
		}),
		ExpectMustValue(func() (Timestamp, error) { return ParseTimestampFromString("2024-01-02") }, Equal(Timestamp(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)))),
		ExpectMustValue(func() (string, error) {
			raw, err := origin.MarshalText()
			return string(raw), err
		}, Equal("2024-01-02T03:04:05Z")),
	)

	var decoded Timestamp
	Then(t, "Timestamp 支持数据库扫描和值转换",
		ExpectDo(func() error { return decoded.Scan(int64(1)) }),
	)
	Then(t, "Timestamp Scan 后可读出 Unix 值",
		Expect(decoded.Unix(), Equal(int64(1))),
		ExpectMustValue(func() (int64, error) {
			v, err := decoded.Value()
			if err != nil {
				return 0, err
			}
			return v.(int64), nil
		}, Equal(int64(1))),
		Expect(decoded.DataType("sqlite"), Equal("bigint")),
		Expect(decoded.OpenAPISchemaFormat(), Equal("date-time")),
	)

	Then(t, "Timestamp 零值判断兼容 nil",
		ExpectDo(func() error { return decoded.Scan(nil) }),
	)
	Then(t, "Timestamp 扫描 nil 后回到零值",
		Expect(decoded.IsZero(), Equal(true)),
	)
}
