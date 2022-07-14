package datatypes

import (
	"database/sql"
	"database/sql/driver"

	"github.com/octohelm/x/encoding"
)

func Scan(b []byte, v any) error {
	if scanner, ok := v.(sql.Scanner); ok {
		return scanner.Scan(b)
	}
	return encoding.UnmarshalText(v, b)
}

func Value(v any) (driver.Value, error) {
	if valuer, ok := v.(driver.Valuer); ok {
		return valuer.Value()
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}
