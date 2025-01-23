package json

import (
	"database/sql/driver"

	"github.com/go-json-experiment/json"
	jsonv1 "github.com/go-json-experiment/json/v1"
)

func ValueOf[T any](v *T) Value[T] {
	return Value[T]{v: v}
}

type Value[T any] struct {
	v *T
}

func (v *Value[T]) OneOf() []any {
	return []any{new(T)}
}

func (v Value[T]) IsZero() bool {
	return v.v == nil
}

func (v *Value[T]) Set(t *T) {
	v.v = t
}

func (v *Value[T]) Get() *T {
	return v.v
}

func (v Value[T]) As(a *T) {
	if v.v != nil {
		*a = *v.v
	}
}

func (v *Value[T]) UnmarshalJSON(data []byte) error {
	t := new(T)
	if err := json.Unmarshal(data, t, jsonv1.OmitEmptyWithLegacyDefinition(true)); err != nil {
		return err
	}
	*v = Value[T]{
		v: t,
	}
	return nil
}

func (v Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.v, jsonv1.OmitEmptyWithLegacyDefinition(true))
}

func (Value[T]) DataType(driverName string) string {
	return "text"
}

func (v Value[T]) Value() (driver.Value, error) {
	return toValue(v.v)
}

func (v *Value[T]) Scan(src any) error {
	x := new(T)
	if err := scanValue(src, x); err != nil {
		return err
	}
	*v = Value[T]{
		v: x,
	}
	return nil
}
