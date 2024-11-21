package filter

import (
	"github.com/go-json-experiment/json"
	"strconv"

	encodingx "github.com/octohelm/x/encoding"
)

func Lit[T comparable](v T) Value[T] {
	return &lit[T]{
		value: v,
	}
}

type lit[T comparable] struct {
	value T
}

func (v lit[T]) Value() T {
	return v.value
}

func (l *lit[T]) PtrValue() *T {
	return &l.value
}

func (l *lit[T]) UnmarshalJSON(b []byte) error {
	v := new(T)
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	*l = lit[T]{value: *v}
	return nil
}

func (l *lit[T]) UmarshalText(b []byte) (err error) {
	if len(b) == 0 {
		return nil
	}
	if b[0] == '"' {
		raw, err := strconv.Unquote(string(b))
		if err != nil {
			return err
		}
		b = []byte(raw)
	}
	v := new(T)
	if err := encodingx.UnmarshalText(v, b); err != nil {
		return err
	}
	*l = lit[T]{value: *v}
	return nil
}

func (v lit[T]) MarshalText() (text []byte, err error) {
	return json.Marshal(v.value)
}

func (l lit[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.value)
}
