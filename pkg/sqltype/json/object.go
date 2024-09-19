package json

import (
	"database/sql/driver"
	"encoding/json"
)

func ObjectOf[T any](data *T) Object[T] {
	return Object[T]{Data: data}
}

type Object[T any] struct {
	Data *T `json:",inline"`
}

func (v *Object[T]) IsZero() bool {
	return v == nil || v.Data == nil
}

func (v *Object[T]) OneOf() []any {
	return []any{new(T)}
}

func (v *Object[T]) Set(t *T) {
	v.Data = t
}

func (v Object[T]) As(a *T) {
	if v.Data != nil {
		*a = *v.Data
	}
}

func (v *Object[T]) UnmarshalJSON(data []byte) error {
	t := new(T)
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	*v = Object[T]{
		Data: t,
	}
	return nil
}

func (v Object[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Data)
}

func (Object[T]) DataType(driverName string) string {
	return "text"
}

func (v Object[T]) Value() (driver.Value, error) {
	return toValue(v.Data)
}

func (v *Object[T]) Scan(src any) error {
	var data T
	if err := scanValue(src, &data); err != nil {
		return err
	}
	*v = Object[T]{
		Data: &data,
	}
	return nil
}
