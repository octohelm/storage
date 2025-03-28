package filter

import (
	"encoding"
	"iter"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
)

type Arg interface {
	encoding.TextMarshaler
}

type Value[T comparable] interface {
	Arg
	Value() T
}

type TypedRule[T comparable] interface {
	Rule

	New() *T
}

type Rule interface {
	Arg
	Op() Op

	IsZero() bool
	Args() iter.Seq[Arg]

	directive.Unmarshaler
}

type RuleExpr interface {
	IsZero() bool
	WhereOf(name string) Rule
}
