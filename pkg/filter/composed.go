package filter

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"reflect"
	"strings"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
	slicesx "github.com/octohelm/x/slices"
)

func Compose(filters ...any) *Composed {
	c := &Composed{}

	for _, filter := range filters {
		rv := reflect.ValueOf(filter)

		if rv.Kind() == reflect.Struct {
			t := rv.Type()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}

			for i := 0; i < rv.NumField(); i++ {
				f := t.Field(i)

				if !ast.IsExported(f.Name) {
					continue
				}

				fv := rv.Field(i)

				if ruleExpr, ok := fv.Interface().(RuleExpr); ok {
					name := f.Name

					if tagName, ok := f.Tag.Lookup("name"); ok {
						n := strings.SplitN(tagName, ",", 2)[0]
						if n != "" {
							name = n
						}
					}

					c.register(name, &fieldRuler{
						name:        name,
						tpe:         t,
						ruleExprIdx: i,
					})

					if fv.Kind() == reflect.Ptr && fv.IsNil() {
						continue
					}

					if !ruleExpr.IsZero() {
						c.rules = append(c.rules, ruleExpr.WhereOf(name))
					}

					break
				}
			}
		}
	}

	return c
}

type fieldRuler struct {
	tpe         reflect.Type
	name        string
	ruleExprIdx int
}

func (t *fieldRuler) Name() string {
	return t.name
}

func (t *fieldRuler) New() *ruleWrapper {
	rv := reflect.New(t.tpe)

	f := rv.Elem().Field(t.ruleExprIdx)
	if f.Kind() == reflect.Ptr {
		f.Set(reflect.New(f.Type().Elem()))
	}

	return &ruleWrapper{
		obj:  rv.Interface(),
		rule: f.Interface().(Rule),
	}
}

type ruleWrapper struct {
	obj  any
	rule Rule
}

func (r *ruleWrapper) Obj() any {
	return r.obj
}

func (r *ruleWrapper) Rule() Rule {
	return r.rule
}

func (r *ruleWrapper) UnmarshalDirective(dec *directive.Decoder) error {
	return r.rule.UnmarshalDirective(dec)
}

type Composed struct {
	Filters []any

	fieldRulers map[string]*fieldRuler
	rules       []Arg
}

func (c *Composed) register(fieldName string, fr *fieldRuler) {
	if c.fieldRulers == nil {
		c.fieldRulers = map[string]*fieldRuler{}
	}
	c.fieldRulers[fieldName] = fr
}

func (c Composed) IsZero() bool {
	return len(c.rules) == 0
}

func (c *Composed) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	d := directive.NewDecoder(bytes.NewBuffer(data))

	d.RegisterDirectiveNewer("or", func() directive.Unmarshaler {
		return directive.UnmarshalerFunc(func(dec *directive.Decoder) error {
			_, err := dec.DirectiveName()
			if err != nil {
				return err
			}

			for {
				k, text := dec.Next()
				if k == directive.EOF || k == directive.KindFuncEnd {
					break
				}

				switch k {
				case directive.KindFuncStart:
					sub, err := dec.Unmarshaler(string(text))
					if err != nil {
						return err
					}
					if err := sub.UnmarshalDirective(dec); err != nil {
						return err
					}
					if arg, ok := sub.(Arg); ok {
						c.rules = append(c.rules, arg)
					}
				default:

				}
			}

			return nil
		})
	})

	d.RegisterDirectiveNewer("where", func() directive.Unmarshaler {
		return directive.UnmarshalerFunc(func(dec *directive.Decoder) error {
			_, err := dec.DirectiveName()
			if err != nil {
				return err
			}

			argIdx := 0
			var fr *fieldRuler

			for {
				k, text := dec.Next()
				if k == directive.EOF || k == directive.KindFuncEnd {
					break
				}

				switch k {
				case directive.KindValue:
					if argIdx == 0 {
						name := ""
						if err := json.Unmarshal(text, &name); err != nil {
							return err
						}

						n, ok := c.fieldRulers[name]
						if ok {
							fr = n
							continue
						}

						return &ErrUnsupportedQLField{
							FieldName: name,
						}
					}
					argIdx++
				case directive.KindFuncStart:
					if fr == nil {
						return &directive.ErrInvalidDirective{}
					}

					wrapper := fr.New()

					if err := wrapper.UnmarshalDirective(dec); err != nil {
						return err
					}

					if ruleExpr, ok := wrapper.rule.(RuleExpr); ok {
						c.rules = append(c.rules, ruleExpr.WhereOf(fr.Name()))
					}

					c.Filters = append(c.Filters, wrapper.obj)

					argIdx++
				default:

				}
			}

			return nil
		})
	})

	_, err := d.DirectiveName()
	if err != nil {
		return err
	}

	for {
		k, text := d.Next()
		if k == directive.EOF || k == directive.KindFuncEnd {
			break
		}

		switch k {
		case directive.KindFuncStart:
			sub, err := d.Unmarshaler(string(text))
			if err != nil {
				return err
			}
			if err := sub.UnmarshalDirective(d); err != nil {
				return err
			}
		default:

		}
	}

	return nil
}

func (c Composed) MarshalText() ([]byte, error) {
	switch len(c.rules) {
	case 0:
		return nil, nil
	case 1:
		return c.rules[0].MarshalText()
	}
	return directive.MarshalDirective("or", slicesx.Map(c.rules, func(e Arg) any {
		return e
	})...)
}
