package filter

import (
	"encoding"
	"iter"

	"github.com/octohelm/storage/pkg/filter/internal/directive"
)

// Arg 表示可序列化的过滤参数。
type Arg interface {
	encoding.TextMarshaler
}

// Value 表示持有具体值的过滤参数。
type Value[T comparable] interface {
	Arg
	Value() T
}

// TypedRule 表示带值类型信息的过滤规则。
type TypedRule[T comparable] interface {
	Rule

	New() *T
}

// Rule 表示通用过滤规则。
type Rule interface {
	Arg
	Op() Op

	IsZero() bool
	Args() iter.Seq[Arg]

	directive.Unmarshaler
}

// RuleExpr 表示可按字段名展开为 Rule 的规则表达式。
type RuleExpr interface {
	IsZero() bool
	WhereOf(name string) Rule
}
