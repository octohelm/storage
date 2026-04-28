package time

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	timeUnits = [][]string{
		{"ns", "nano"},
		{"us", "µs", "micro"},
		{"ms", "milli"},
		{"s", "sec"},
		{"m", "min"},
		{"h", "hr", "hour"},
		{"d", "day"},
		{"w", "wk", "week"},
	}

	timeMultiplier = map[string]time.Duration{
		"ns": time.Nanosecond,
		"us": time.Microsecond,
		"ms": time.Millisecond,
		"s":  time.Second,
		"m":  time.Minute,
		"h":  time.Hour,
		"d":  24 * time.Hour,
		"w":  7 * 24 * time.Hour,
	}

	durationMatcher = regexp.MustCompile(`((\d+)\s*([A-Za-zµ]+))`)
)

// IsDuration 判断字符串能否被解析为时长。
func IsDuration(str string) bool {
	_, err := ParseDuration(str)
	return err == nil
}

// Duration 表示可序列化、可扫描的时长值。
type Duration time.Duration

// IsZero 判断时长是否为零。
func (d Duration) IsZero() bool {
	return d == 0
}

// OpenAPISchemaFormat 返回 OpenAPI 使用的格式名。
func (Duration) OpenAPISchemaFormat() string {
	return "duration"
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error { // validation is performed later on
	dd, err := ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

// ParseDuration 解析兼容标准格式与扩展单位写法的时长字符串。
func ParseDuration(cand string) (time.Duration, error) {
	if dur, err := time.ParseDuration(cand); err == nil {
		return dur, nil
	}

	var dur time.Duration
	ok := false
	for _, match := range durationMatcher.FindAllStringSubmatch(cand, -1) {

		factor, err := strconv.Atoi(match[2]) // converts string to int
		if err != nil {
			return 0, err
		}
		unit := strings.ToLower(strings.TrimSpace(match[3]))

		for _, variants := range timeUnits {
			last := len(variants) - 1
			multiplier := timeMultiplier[variants[0]]

			for i, variant := range variants {
				if (last == i && strings.HasPrefix(unit, variant)) || strings.EqualFold(variant, unit) {
					ok = true
					dur += time.Duration(factor) * multiplier
				}
			}
		}
	}

	if ok {
		return dur, nil
	}
	return 0, fmt.Errorf("unable to parse %s as duration", cand)
}

func (d *Duration) Scan(raw any) error {
	switch v := raw.(type) {
	case int64:
		*d = Duration(v)
	case float64:
		*d = Duration(int64(v))
	case nil:
		*d = Duration(0)
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Duration from: %#v", v)
	}

	return nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}
