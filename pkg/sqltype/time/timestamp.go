package time

import (
	"cmp"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strconv"
	"time"
)

import (
	_ "time/tzdata"
)

var (
	UTC               = time.UTC
	TimestampZero     = Timestamp(time.Time{})
	TimestampUnixZero = Timestamp(time.Unix(0, 0))
)

// CST 表示默认使用的时区位置。
var CST = time.UTC

func init() {
	cst, _ := time.LoadLocation("Asia/Shanghai")
	if cst != nil {
		CST = cst
	}
}

var (
	outputLayout     = time.RFC3339
	supportedLayouts = map[string]*time.Location{
		time.RFC3339: CST,
	}
)

// AddSupportedLayout 注册一个可解析的时间布局。
func AddSupportedLayout(layout string, location *time.Location) {
	supportedLayouts[layout] = cmp.Or(location, CST)
}

// SetOutputLayout 设置时间输出布局，并可同步更新默认时区。
func SetOutputLayout(layout string, location *time.Location) {
	outputLayout = layout

	if location != nil {
		CST = location
	}
}

// Now 返回当前时间对应的 Timestamp。
func Now() Timestamp {
	return Timestamp(time.Now())
}

// Add 返回时间加上指定时长后的结果。
func Add(t Timestamp, d time.Duration) Timestamp {
	return Timestamp(time.Time(t).Add(d))
}

// Sub 返回两个时间之间的时长差。
func Sub(t Timestamp, u Timestamp) time.Duration {
	return time.Time(t).Sub(time.Time(u))
}

// AddDate 返回时间按年月日偏移后的结果。
func AddDate(t Timestamp, years int, months int, days int) Timestamp {
	return Timestamp(time.Time(t).AddDate(years, months, days))
}

// Timestamp 表示以秒级 Unix 时间存储的时间值。
type Timestamp time.Time

// OpenAPISchemaFormat 返回 OpenAPI 使用的格式名。
func (Timestamp) OpenAPISchemaFormat() string {
	return "date-time"
}

// DataType 返回数据库侧的数据类型名。
func (Timestamp) DataType(engine string) string {
	return "bigint"
}

// ParseTimestampFromString 按已注册布局解析时间字符串。
func ParseTimestampFromString(s string) (d Timestamp, err error) {
	for layout, cst := range supportedLayouts {
		t, e := time.ParseInLocation(layout, s, cst)
		if e == nil {
			return Timestamp(t), nil
		}
		err = e
	}
	return d, err
}

// ParseTimestampFromStringWithLayout 按指定布局解析时间字符串。
func ParseTimestampFromStringWithLayout(input, layout string) (Timestamp, error) {
	t, err := time.ParseInLocation(layout, input, CST)
	if err != nil {
		return TimestampUnixZero, err
	}
	return Timestamp(t), nil
}

var _ interface {
	sql.Scanner
	driver.Valuer
} = (*Timestamp)(nil)

func (dt *Timestamp) Scan(value any) error {
	switch x := value.(type) {
	case []byte:
		v, err := strconv.ParseInt(string(x), 10, 64)
		if err != nil {
			return fmt.Errorf("sql.Scan() strfmt.Timestamp from: %#v failed: %w", v, err)
		}
		if v < 0 {
			v = 0
		}
		*dt = Timestamp(time.Unix(v, 0))
	case int64:
		if x < 0 {
			x = 0
		}
		*dt = Timestamp(time.Unix(x, 0))
	case nil:
		*dt = TimestampZero
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Timestamp from: %v", x)
	}
	return nil
}

func (dt Timestamp) Value() (driver.Value, error) {
	s := max(time.Time(dt).Unix(), 0)
	return s, nil
}

func (dt Timestamp) String() string {
	if dt.IsZero() {
		return ""
	}
	return time.Time(dt).In(CST).Format(outputLayout)
}

func (dt Timestamp) Format(layout string) string {
	return time.Time(dt).In(CST).Format(layout)
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*Timestamp)(nil)

func (dt Timestamp) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *Timestamp) UnmarshalText(data []byte) (err error) {
	str := string(data)
	if len(str) == 0 || str == "0" {
		return nil
	}
	*dt, err = ParseTimestampFromString(str)
	return err
}

func (dt Timestamp) IsZero() bool {
	unix := dt.Unix()
	return unix == 0 || unix == TimestampZero.Unix()
}

func (dt Timestamp) Unix() int64                  { return time.Time(dt).Unix() }
func (dt Timestamp) Year() int                    { return time.Time(dt).Year() }
func (dt Timestamp) Month() time.Month            { return time.Time(dt).Month() }
func (dt Timestamp) Day() int                     { return time.Time(dt).Day() }
func (dt Timestamp) Date() (int, time.Month, int) { return time.Time(dt).Date() }
func (dt Timestamp) In(loc *time.Location) Timestamp {
	return Timestamp(time.Time(dt).In(loc))
}
