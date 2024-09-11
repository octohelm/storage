package sqlbuilder

import (
	"context"
	"iter"
	"sort"

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
	AdditionLimit
	AdditionOnConflict
	AdditionOther
	AdditionComment
)

type Additions []Addition

func (additions Additions) Len() int {
	return len(additions)
}

func (additions Additions) Less(i, j int) bool {
	return additions[i].AdditionType() < additions[j].AdditionType()
}

func (additions Additions) Swap(i, j int) {
	additions[i], additions[j] = additions[j], additions[i]
}

func (additions Additions) IsNil() bool {
	return len(additions) == 0
}

func (additions Additions) Frag(ctx context.Context) iter.Seq2[string, []any] {
	sort.Sort(additions)

	return func(yield func(string, []any) bool) {
		for _, add := range additions {
			if sqlfrag.IsNil(add) {
				continue
			}

			if !yield("\n", nil) {
				return
			}

			if !yield(sqlfrag.All(ctx, add)) {
				return
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
