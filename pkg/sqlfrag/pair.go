package sqlfrag

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"text/scanner"
)

func Pair(query string, args ...any) Fragment {
	if len(args) == 0 {
		return Const(query)
	}

	namedArgSet := NamedArgSet{}
	finalArgs := make([]any, 0, len(args))

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
	args        []any
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
						for q, args := range iterArg(ctx, arg) {
							if !yield(q, args) {
								return
							}
						}
					} else {
						panic(fmt.Errorf("missing named arg `%s`", name))
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
					panic(fmt.Errorf("missing arg %d of %s", argIndex, p.query))
				}

				for q, args := range iterArg(ctx, arg) {
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
