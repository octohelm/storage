package text

import (
	"database/sql/driver"
)

type Text string

func (v *Text) Set(str string) {
	*v = Text(str)
}

func (Text) DataType(driverName string) string {
	return "text"
}

func (v Text) Value() (driver.Value, error) {
	if v == "" {
		return nil, nil
	}
	return string(v), nil
}

func (v *Text) Scan(src any) error {
	switch x := src.(type) {
	case string:
		*v = Text(x)
	case []byte:
		*v = Text(x)
	}
	return nil
}
