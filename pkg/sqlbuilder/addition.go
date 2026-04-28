package sqlbuilder

import (
	"cmp"
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

// Addition 表示可挂载到语句上的附加子句。
type Addition interface {
	sqlfrag.Fragment

	AdditionType() AdditionType
}

// AdditionType 表示附加子句的种类。
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
	AdditionLock
	AdditionOther
	AdditionComment
)

// Additions 表示一组附加子句。
type Additions []Addition

// CompareAddition 按附加子句类型排序。
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

// AsAddition 把任意片段包装为指定类型的附加子句。
func AsAddition(additionType AdditionType, fragment sqlfrag.Fragment) Addition {
	return &addition{
		additionType: additionType,
		Fragment:     fragment,
	}
}

type addition struct {
	sqlfrag.Fragment

	additionType AdditionType
}

func (a *addition) IsNil() bool {
	return a == nil || sqlfrag.IsNil(a.Fragment)
}

func (a *addition) AdditionType() AdditionType {
	return a.additionType
}
