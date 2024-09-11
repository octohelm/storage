package sqlfrag

import (
	"bytes"
	"context"
	"database/sql/driver"
	"iter"
	"reflect"
	"text/scanner"

	reflectx "github.com/octohelm/x/reflect"
	"github.com/pkg/errors"
)

func Pair(query string, args ...any) Fragment {
	if len(args) == 0 {
		return Const(query)
	}

	namedArgSet := NamedArgSet{}
	finalArgs := make(Values, 0, len(args))

	for _, arg := range args {
		switch x := arg.(type) {
		case NamedArgSet:
			for k := range x {
				namedArgSet[k] = x[k]
			}
		case NamedArg:
			namedArgSet[x.Name] = x.Value
		default:
			finalArgs = append(finalArgs, x)
		}
	}

	return &pair{
		query:       query,
		args:        finalArgs,
		namedArgSet: namedArgSet,
	}
}

type pair struct {
	query       string
	args        Values
	namedArgSet NamedArgSet
}

func (p *pair) IsNil() bool {
	return p.query == "" && len(p.args) == 0
}

func (p *pair) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		s := &scanner.Scanner{}
		s.Init(bytes.NewBuffer([]byte(p.query)))
		s.Error = func(s *scanner.Scanner, msg string) {}

		argIndex := 0
		tmp := bytes.NewBuffer(nil)

		for c := s.Next(); c != scanner.EOF; c = s.Next() {
			switch c {
			case '@':
				if tmp.Len() > 0 {
					if !yield(tmp.String(), nil) {
						return
					}
				}

				tmp.Reset()
				named := bytes.NewBuffer(nil)

				for {
					c = s.Next()

					if c == scanner.EOF {
						break
					}

					if (c >= 'A' && c <= 'Z') ||
						(c >= 'a' && c <= 'z') ||
						(c >= '0' && c <= '9') ||
						c == '_' {

						named.WriteRune(c)
						continue
					}

					tmp.WriteRune(c)
					break
				}

				if named.Len() > 0 {
					name := named.String()
					if arg, ok := p.get(name); ok {
						for q, args := range p.pairSeq(ctx, arg) {
							if !yield(q, args) {
								return
							}
						}
					} else {
						panic(errors.Errorf("missing named arg `%s`", name))
					}
				}
			case '?':
				if tmp.Len() > 0 {
					if !yield(tmp.String(), nil) {
						return
					}
				}

				tmp.Reset()

				arg, ok := p.arg(argIndex)
				if !ok {
					panic(errors.Errorf("missing arg %d of %s", argIndex, p.query))
				}
				for q, args := range p.pairSeq(ctx, arg) {
					if !yield(q, args) {
						return
					}
				}
				argIndex++
			default:
				tmp.WriteRune(c)
			}
		}

		if tmp.Len() > 0 {
			if !yield(tmp.String(), nil) {
				return
			}
		}
	}
}

func (p *pair) pairSeq(ctx context.Context, v any) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		switch arg := v.(type) {
		case CustomValueArg:
			if !yield(arg.ValueEx(), Values{arg}) {
				return
			}
		case Fragment:
			if !IsNil(arg) {
				for q, args := range arg.Frag(ctx) {
					if !yield(q, args) {
						return
					}
				}
			}
		case driver.Valuer:
			if !yield("?", Values{arg}) {
				return
			}
		case []any:
			if values := Values(arg); !IsNil(values) {
				for q, args := range values.Frag(ctx) {
					if !yield(q, args) {
						return
					}
				}
			}
		default:
			if typ := reflect.TypeOf(arg); typ.Kind() == reflect.Slice {
				if !reflectx.IsBytes(typ) {
					if values := toValues(arg); !IsNil(values) {
						for q, args := range values.Frag(ctx) {
							if !yield(q, args) {
								return
							}
						}
					}
					return
				}
			}

			if !yield("?", Values{arg}) {
				return
			}
		}
	}
}

func (p *pair) get(name string) (any, bool) {
	if v, ok := p.namedArgSet[name]; ok {
		return v, true
	}
	return nil, false
}

func (p *pair) arg(i int) (any, bool) {
	if i < len(p.args) {
		return p.args[i], true
	}
	return nil, false
}

func toValues(arg any) Values {
	switch x := (arg).(type) {
	case []bool:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []string:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []float32:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []float64:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int8:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int16:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int32:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []int64:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint8:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint16:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint32:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []uint64:
		values := make(Values, len(x))
		for i := range values {
			values[i] = x[i]
		}
		return values
	case []any:
		return x
	}
	sliceRv := reflect.ValueOf(arg)
	values := make(Values, sliceRv.Len())
	for i := range values {
		values[i] = sliceRv.Index(i).Interface()
	}
	return values
}
