package sqlfrag

import (
	"context"
	"database/sql"
	"iter"
	"strings"
)

type NamedArg = sql.NamedArg

type NamedArgSet map[string]any

// CustomValueArg
// replace ? as some query snippet
//
// examples:
// ? => ST_GeomFromText(?)
type CustomValueArg interface {
	ValueEx() string
}

type Values []any

var _ Fragment = Values{}

func (v Values) IsNil() bool {
	return len(v) == 0
}

func (v Values) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		for q, args := range Pair(strings.Repeat(",?", len(v))[1:], v...).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
