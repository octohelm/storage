package filter

import (
	slicesx "github.com/octohelm/x/slices"
)

// +gengo:enum
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

	OP__WHERE
	OP__AND
	OP__OR
)

// Eq == v
func Eq[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__EQ,
		args: []Arg{
			Lit(v),
		},
	}
}

// Neq != v
func Neq[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__NEQ,
		args: []Arg{
			Lit(v),
		},
	}
}

// Lt < v
func Lt[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__LT,
		args: []Arg{
			Lit(v),
		},
	}
}

// Lte <= v
func Lte[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__LTE,
		args: []Arg{
			Lit(v),
		},
	}
}

// In values
func In[T comparable](values []T) *Filter[T] {
	return &Filter[T]{
		op: OP__IN,
		args: slicesx.Map(values, func(e T) Arg {
			return Lit(e)
		}),
	}
}

// Notin values
func Notin[T comparable](values []T) *Filter[T] {
	return &Filter[T]{
		op: OP__NOTIN,
		args: slicesx.Map(values, func(e T) Arg {
			return Lit(e)
		}),
	}
}

// Gt > v
func Gt[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__GT,
		args: []Arg{
			Lit(v),
		},
	}
}

// Gte >= v
func Gte[T comparable](v T) *Filter[T] {
	return &Filter[T]{
		op: OP__GTE,
		args: []Arg{
			Lit(v),
		},
	}
}

func And[T comparable](filters ...TypedRule[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__AND,
		args: slicesx.Map(filters, func(f TypedRule[T]) Arg {
			return f
		}),
	}
}

func Or[T comparable](filters ...TypedRule[T]) *Filter[T] {
	return &Filter[T]{
		op: OP__OR,
		args: slicesx.Map(filters, func(f TypedRule[T]) Arg {
			return f
		}),
	}
}

func OrRules(rules ...Rule) Rule {
	return &Filter[any]{
		op: OP__OR,
		args: slicesx.Map(rules, func(e Rule) Arg {
			return e
		}),
	}
}
