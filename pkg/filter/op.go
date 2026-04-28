package filter

import (
	"iter"
	"slices"

	slicesx "github.com/octohelm/x/slices"

	"github.com/octohelm/storage/internal/xiter"
)

// +gengo:enum
// Op 表示过滤条件的操作符类型。
type Op uint8

const (
	OP_UNKNOWN Op = iota

	OP__EQ
	OP__NEQ
	OP__IN
	OP__NOTIN

	OP__GTE
	OP__GT
	OP__LTE
	OP__LT

	OP__NOTCONTAINS
	OP__CONTAINS
	OP__PREFIX
	OP__SUFFIX

	OP__WHERE
	OP__AND
	OP__OR

	OP__INTERSECTION
)

// Eq 构造等于指定值的过滤条件。
func Eq[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__EQ,
		args: []Arg{
			Lit(v),
		},
	}
}

// Neq 构造不等于指定值的过滤条件。
func Neq[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__NEQ,
		args: []Arg{
			Lit(v),
		},
	}
}

// Lt 构造小于指定值的过滤条件。
func Lt[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__LT,
		args: []Arg{
			Lit(v),
		},
	}
}

// Lte 构造小于等于指定值的过滤条件。
func Lte[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__LTE,
		args: []Arg{
			Lit(v),
		},
	}
}

// Contains 构造包含指定值的过滤条件。
func Contains[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__CONTAINS,
		args: []Arg{
			Lit(v),
		},
	}
}

// Prefix 构造前缀匹配的过滤条件。
func Prefix[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__PREFIX,
		args: []Arg{
			Lit(v),
		},
	}
}

// Suffix 构造后缀匹配的过滤条件。
func Suffix[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__SUFFIX,
		args: []Arg{
			Lit(v),
		},
	}
}

// In 构造值属于给定集合的过滤条件。
func In[T comparable](values ...T) *Filter[T] {
	return &Filter[T]{
		op: OP__IN,
		args: slicesx.Map(values, func(e T) Arg {
			return Lit(e)
		}),
	}
}

// InSeq 构造值属于给定序列的过滤条件。
func InSeq[T comparable](values iter.Seq[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__IN,
		args: slices.Collect(xiter.Map(values, func(e T) Arg {
			return Lit(e)
		})),
	}
}

// Notin 构造值不属于给定集合的过滤条件。
func Notin[T comparable](values ...T) *Filter[T] {
	return &Filter[T]{
		op: OP__NOTIN,
		args: slicesx.Map(values, func(e T) Arg {
			return Lit(e)
		}),
	}
}

// NotinSeq 构造值不属于给定序列的过滤条件。
func NotinSeq[T comparable](values iter.Seq[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__NOTIN,
		args: slices.Collect(xiter.Map(values, func(e T) Arg {
			return Lit(e)
		})),
	}
}

// Gt 构造大于指定值的过滤条件。
func Gt[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__GT,
		args: []Arg{
			Lit(v),
		},
	}
}

// Gte 构造大于等于指定值的过滤条件。
func Gte[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__GTE,
		args: []Arg{
			Lit(v),
		},
	}
}

// And 构造多个条件的与组合。
func And[T comparable](filters ...TypedRule[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__AND,
		args: slicesx.Map(filters, func(f TypedRule[T]) Arg {
			return f
		}),
	}
}

// Or 构造多个条件的或组合。
func Or[T comparable](filters ...TypedRule[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__OR,
		args: slicesx.Map(filters, func(f TypedRule[T]) Arg {
			return f
		}),
	}
}

// Intersection 构造多个条件的交集组合。
func Intersection[T comparable](filters ...TypedRule[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__INTERSECTION,
		args: slicesx.Map(filters, func(f TypedRule[T]) Arg {
			return f
		}),
	}
}

// OrRules 用 Rule 构造或组合过滤条件。
func OrRules(rules ...Rule) Rule {
	return &Filter[any]{
		op: OP__OR,
		args: slicesx.Map(rules, func(e Rule) Arg {
			return e
		}),
	}
}
