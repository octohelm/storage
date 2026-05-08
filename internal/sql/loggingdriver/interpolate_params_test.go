package loggingdriver

import (
	"database/sql/driver"
	"testing"
	"time"

	. "github.com/octohelm/x/testing/v2"
)

func TestInterpolateParams(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	ts := time.Date(2024, 1, 2, 3, 4, 5, 123456000, time.UTC)

	sql, err := InterpolateParams("SELECT ?, ?, ?, ?, ?, ?, ?", []driver.NamedValue{
		{Ordinal: 1, Value: int64(1)},
		{Ordinal: 2, Value: float64(1.5)},
		{Ordinal: 3, Value: true},
		{Ordinal: 4, Value: "a\nb"},
		{Ordinal: 5, Value: []byte("x'y")},
		{Ordinal: 6, Value: ts},
		{Ordinal: 7, Value: nil},
	}, loc)

	Then(
		t, "InterpolateParams 覆盖基础参数类型",
		Expect(err, Equal(error(nil))),
		Expect(sql, Equal("SELECT 1, 1.5, 1, 'a\\nb', E'x\\'y', '2024-01-02 11:04:05.123456', NULL")),
	)

	zeroTimeSQL, err := InterpolateParams("SELECT ?", []driver.NamedValue{{Ordinal: 1, Value: time.Time{}}}, time.UTC)
	Then(
		t, "零值时间格式化为固定日期",
		Expect(err, Equal(error(nil))),
		Expect(zeroTimeSQL, Equal("SELECT '0000-00-00'")),
	)

	Then(
		t, "占位符数量不匹配返回 ErrSkip",
		ExpectDo(func() error {
			_, err := InterpolateParams("SELECT ?", nil, time.UTC)
			return err
		}, ErrorIs(driver.ErrSkip)),
	)

	Then(
		t, "不支持的参数类型返回错误",
		ExpectDo(func() error {
			_, err := InterpolateParams("SELECT ?", []driver.NamedValue{{Ordinal: 1, Value: struct{}{}}}, time.UTC)
			return err
		}, ErrorMatch(mustRegexp("unsupported type"))),
	)
}

func TestSqlPrinterAndBuffers(t *testing.T) {
	printer := interpolateParams("SELECT\t?", []driver.NamedValue{{Ordinal: 1, Value: "ok"}})
	Then(
		t, "SqlPrinter.String 会展开 tab 并格式化参数",
		Expect(printer.String(), Equal("SELECT    'ok'")),
	)

	invalidPrinter := interpolateParams("SELECT ?", nil)
	Then(
		t, "非法参数会在 String 中暴露 invalid 前缀",
		Expect(invalidPrinter.String(), Equal("invalid: driver: skip fast-path; continue as if unimplemented")),
	)

	escaped := string(escapeBytesBackslash(nil, []byte{'\x00', '\n', '\r', '\x1a', '\'', '"', '\\', 'a'}))
	Then(
		t, "escapeBytesBackslash 处理特殊字符",
		Expect(escaped, Equal("\\0\\n\\r\\Z\\'\\\"\\\\a")),
	)

	buf := reserveBuffer(make([]byte, 2, 2), 10)
	Then(
		t, "reserveBuffer 会扩容到可写长度",
		Expect(len(buf), Equal(12)),
		Expect(cap(buf) >= 12, Equal(true)),
	)
}
