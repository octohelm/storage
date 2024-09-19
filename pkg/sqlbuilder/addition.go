package sqlbuilder

import (
	"cmp"
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type Addition interface {
	sqlfrag.Fragment

	AdditionType() AdditionType
}

type AdditionType int

const (
	AdditionJoin AdditionType = iota
	AdditionWhere
	AdditionGroupBy
	AdditionCombination
	AdditionOrderBy
	AdditionOnConflict
	AdditionReturning
	AdditionLimit
	AdditionOther
	AdditionComment
)

type Additions []Addition

func CompareAddition(a Addition, b Addition) int {
	return cmp.Compare(a.AdditionType(), b.AdditionType())
}

func (additions Additions) IsNil() bool {
	return len(additions) == 0
}

func (additions Additions) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		for _, add := range slices.SortedFunc(slices.Values(additions), CompareAddition) {
			if sqlfrag.IsNil(add) {
				continue
			}

			if !yield("\n", nil) {
				return
			}

			for q, args := range add.Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
		}
	}
}

func AsAddition(fragment sqlfrag.Fragment) *OtherAddition {
	return &OtherAddition{
		Fragment: fragment,
	}
}

type OtherAddition struct {
	sqlfrag.Fragment
}

func (a *OtherAddition) IsNil() bool {
	return a == nil || sqlfrag.IsNil(a.Fragment)
}

func (OtherAddition) AdditionType() AdditionType {
	return AdditionOther
}
