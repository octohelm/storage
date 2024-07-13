package filter

import (
	"bytes"
	"strings"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
	slicesx "github.com/octohelm/x/slices"
)

type Filter[T comparable] struct {
	op   Op
	args []Arg
}

func (v Filter[T]) New() *T {
	return new(T)
}

func (f Filter[T]) WhereOf(name string) Rule {
	return Where[T](name, &f)
}

func (f Filter[T]) MarshalDirective() ([]byte, error) {
	if f.IsZero() {
		return nil, nil
	}

	return directive.MarshalDirective(strings.ToLower(f.op.String()), slicesx.Map(f.args, func(e Arg) any {
		return e
	})...)
}

func (Filter[T]) OneOf() []any {
	return []any{new(T)}
}

func (f Filter[T]) Op() Op {
	return f.op
}

func (v Filter[T]) Args() []Arg {
	return v.args
}

func (f Filter[T]) IsZero() bool {
	return f.op == OP_UNKNOWN || len(f.args) == 0
}

func (f *Filter[T]) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	ff := Filter[T]{}

	dec := directive.NewDecoder(bytes.NewBuffer(data))

	dec.RegisterDirectiveNewer(directive.DefaultDirectiveNewer, func() directive.Unmarshaler {
		return &Filter[T]{}
	})

	kind, _ := dec.Next()
	if kind != directive.KindFuncStart {
		v := lit[T]{}

		if err := v.UmarshalText(data); err != nil {
			return err
		}

		ff.op = OP__EQ
		ff.args = append(ff.args, v)
	} else {
		err := ff.UnmarshalDirective(dec)
		if err != nil {
			return err
		}
	}

	*f = ff

	return nil
}

func (f *Filter[T]) UnmarshalDirective(dec *directive.Decoder) error {
	name, err := dec.DirectiveName()
	if err != nil {
		return err
	}

	ff := Filter[T]{}
	if err := ff.op.UnmarshalText([]byte(strings.ToUpper(name))); err != nil {
		return err
	}

	for {
		k, text := dec.Next()
		if k == directive.EOF || k == directive.KindFuncEnd {
			break
		}

		switch k {
		case directive.KindValue:
			l := &lit[T]{}
			if err := l.UnmarshalJSON(text); err != nil {
				return err
			}
			ff.args = append(ff.args, l)
		case directive.KindFuncStart:
			sub, err := dec.Unmarshaler(string(text))
			if err != nil {
				return err
			}
			if err := sub.UnmarshalDirective(dec); err != nil {
				return err
			}
			if arg, ok := sub.(Arg); ok {
				ff.args = append(ff.args, arg)
			}
		default:

		}
	}

	*f = ff

	return nil
}

func (f Filter[T]) MarshalText() ([]byte, error) {
	return f.MarshalDirective()
}

func (f Filter[T]) String() string {
	txt, _ := f.MarshalText()
	return string(txt)
}
