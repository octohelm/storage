package time

import (
	"testing"
	"time"

	. "github.com/octohelm/x/testing/v2"
)

func TestDurationHelpers(t *testing.T) {
	Then(
		t, "ParseDuration 支持扩展单位",
		ExpectMustValue(func() (time.Duration, error) { return ParseDuration("1day 2hour 3min") }, Equal(26*time.Hour+3*time.Minute)),
		Expect(IsDuration("1wk"), Equal(true)),
		Expect(IsDuration("bad"), Equal(false)),
	)

	var d Duration
	Then(
		t, "Duration 支持文本与数据库扫描",
		ExpectDo(func() error { return d.UnmarshalText([]byte("2h")) }),
	)
	Then(
		t, "Duration 文本反序列化后可格式化",
		Expect(d.String(), Equal("2h0m0s")),
		Expect(d.OpenAPISchemaFormat(), Equal("duration")),
	)
	Then(
		t, "Duration 支持数据库扫描",
		ExpectDo(func() error { return d.Scan(float64(3)) }),
	)
	Then(
		t, "Duration Scan 后非零",
		Expect(d.IsZero(), Equal(false)),
	)
}
