package text

import (
	"database/sql/driver"
	"strings"
)

type Blob []byte

func (v *Blob) Set(str []byte) {
	*v = Blob(str)
}

func (Blob) DataType(driverName string) string {
	if strings.HasPrefix(driverName, "postgres") {
		return "bytea"
	}
	return "blob"
}

func (v Blob) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	return []byte(v), nil
}

func (v *Blob) Scan(src any) error {
	switch x := src.(type) {
	case string:
		*v = Blob(x)
	case []byte:
		*v = Blob(x)
	}
	return nil
}
