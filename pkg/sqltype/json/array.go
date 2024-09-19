package json

import (
	"database/sql/driver"
)

type Array[T any] []T

func (Array[T]) DataType(driverName string) string {
	return "text"
}

func (v Array[T]) Value() (driver.Value, error) {
	return toValue(v)
}

func (v *Array[T]) Scan(src any) error {
	return scanValue(src, v)
}
