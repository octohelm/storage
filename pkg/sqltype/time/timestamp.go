package time

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strconv"
	"time"
	_ "time/tzdata"
)

var (
	UTC               = time.UTC
	TimestampZero     = Timestamp(time.Time{})
	TimestampUnixZero = Timestamp(time.Unix(0, 0))
)

var (
	CST          = time.UTC
	OutputLayout = time.RFC3339
)

func init() {
	cst, _ := time.LoadLocation("Asia/Shanghai")
	if cst != nil {
		CST = cst
	}
}

func SetOutput(layout string, location *time.Location) {
	OutputLayout = layout

	if location != nil {
		CST = location
	}
}

func Now() Timestamp {
	return Timestamp(time.Now())
}

func Add(t Timestamp, d time.Duration) Timestamp {
	return Timestamp(time.Time(t).Add(d))
}

func Sub(t Timestamp, u Timestamp) time.Duration {
	return time.Time(t).Sub(time.Time(u))
}

func AddDate(t Timestamp, years int, months int, days int) Timestamp {
	return Timestamp(time.Time(t).AddDate(years, months, days))
}

type Timestamp time.Time

func (Timestamp) OpenAPISchemaFormat() string {
	return "date-time"
}

func (Timestamp) DataType(engine string) string {
	return "bigint"
}

func ParseTimestampFromString(s string) (Timestamp, error) {
	if OutputLayout != time.RFC3339 {
		t, err := time.ParseInLocation(OutputLayout, s, CST)
		if err == nil {
			return Timestamp(t), nil
		}
	}
	// fallback
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return TimestampUnixZero, err
	}
	return Timestamp(t), nil
}

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

func (dt *Timestamp) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return fmt.Errorf("sql.Scan() strfmt.Timestamp from: %#v failed: %s", v, err.Error())
		}
		*dt = Timestamp(time.Unix(n, 0))
	case int64:
		if v < 0 {
			*dt = Timestamp{}
		} else {
			*dt = Timestamp(time.Unix(v, 0))
		}
	case nil:
		*dt = TimestampZero
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.Timestamp from: %#v", v)
	}
	return nil
}

func (dt Timestamp) Value() (driver.Value, error) {
	s := (time.Time)(dt).Unix()
	if s < 0 {
		s = 0
	}
	return s, nil
}

func (dt Timestamp) String() string {
	if dt.IsZero() {
		return ""
	}
	return time.Time(dt).In(CST).Format(OutputLayout)
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
	return
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
