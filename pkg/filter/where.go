package filter

import (
	"bytes"
	"github.com/go-json-experiment/json"
	"slices"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
	slicesx "github.com/octohelm/x/slices"
)

func Where[T comparable](name string, rules ...TypedRule[T]) TypedRule[T] {
	return &where[T]{
		name: name,
		args: slicesx.Map(rules, func(e TypedRule[T]) Arg {
			return e
		}),
	}
}

type where[T comparable] struct {
	name string
	args []Arg
}

func (w where[T]) Args() []Arg {
	return w.args
}

func (w where[T]) IsZero() bool {
	return len(w.args) == 0
}

func (where[T]) Op() Op {
	return OP__WHERE
}

func (where[T]) New() *T {
	return new(T)
}

func (w *where[T]) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	d := directive.NewDecoder(bytes.NewBuffer(data))

	d.RegisterDirectiveNewer(directive.DefaultDirectiveNewer, func() directive.Unmarshaler {
		return &Filter[T]{}
	})

	return w.UnmarshalDirective(d)
}

func (w *where[T]) UnmarshalDirective(dec *directive.Decoder) error {
	name, err := dec.DirectiveName()
	if err != nil {
		return err
	}
	if name != "where" {
		return &directive.ErrInvalidDirective{}
	}

	dd := &where[T]{}

	argIdx := 0

	for {
		k, text := dec.Next()
		if k == directive.EOF || k == directive.KindFuncEnd {
			break
		}

		switch k {
		case directive.KindValue:
			if argIdx == 0 {
				if err := json.Unmarshal(text, &dd.name); err != nil {
					return err
				}
			}
			argIdx++
		case directive.KindFuncStart:
			sub, err := dec.Unmarshaler(string(text))
			if err != nil {
				return err
			}
			if err := sub.UnmarshalDirective(dec); err != nil {
				return err
			}
			if arg, ok := sub.(Arg); ok {
				dd.args = append(dd.args, arg)
			}
			argIdx++
		default:

		}
	}

	*w = *dd

	return nil
}

func (w where[T]) MarshalText() (text []byte, err error) {
	return w.MarshalDirective()
}

func (w where[T]) MarshalDirective() (text []byte, err error) {
	return directive.MarshalDirective("where", slices.Concat(
		[]any{w.name},
		slicesx.Map(w.args, func(e Arg) any {
			return e
		}),
	)...)
}
