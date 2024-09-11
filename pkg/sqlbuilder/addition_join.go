package sqlbuilder

import (
	"context"
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type JoinAddition interface {
	Addition
	On(joinCondition sqlfrag.Fragment) JoinAddition
	Using(joinColumnList ...Column) JoinAddition
}

func Join(table sqlfrag.Fragment, prefixes ...string) JoinAddition {
	return &join{
		prefix: strings.Join(prefixes, " "),
		target: table,
	}
}

func InnerJoin(table sqlfrag.Fragment) JoinAddition {
	return Join(table, "INNER")
}

func LeftJoin(table sqlfrag.Fragment) JoinAddition {
	return Join(table, "LEFT")
}

func RightJoin(table sqlfrag.Fragment) JoinAddition {
	return Join(table, "RIGHT")
}

func FullJoin(table sqlfrag.Fragment) JoinAddition {
	return Join(table, "FULL")
}

func CrossJoin(table sqlfrag.Fragment) JoinAddition {
	return Join(table, "CROSS")
}

type join struct {
	prefix         string
	target         sqlfrag.Fragment
	joinCondition  sqlfrag.Fragment
	joinColumnList []Column
}

func (j join) AdditionType() AdditionType {
	return AdditionJoin
}

func (j join) On(joinCondition sqlfrag.Fragment) JoinAddition {
	j.joinCondition = joinCondition
	return &j
}

func (j join) Using(joinColumnList ...Column) JoinAddition {
	j.joinColumnList = joinColumnList
	return &j
}

func (j *join) IsNil() bool {
	return j == nil || sqlfrag.IsNil(j.target) || (j.prefix != "CROSS" && sqlfrag.IsNil(j.joinCondition) && len(j.joinColumnList) == 0)
}

func (j *join) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		t := "JOIN "
		if j.prefix != "" {
			t = j.prefix + " " + t
		}

		if !yield(t, nil) {
			return
		}

		if !yield(sqlfrag.All(ctx, j.target)) {
			return
		}

		if !(sqlfrag.IsNil(j.joinCondition)) {
			if !yield(" ON ", nil) {
				return
			}
			if !yield(sqlfrag.All(ctx, j.joinCondition)) {
				return
			}
		}

		if len(j.joinColumnList) > 0 {
			if !yield(" USING (", nil) {
				return
			}

			ctx = ContextWithToggles(ctx, Toggles{
				ToggleMultiTable: false,
			})

			for q, args := range sqlfrag.Join(", ", sqlfrag.NonNil(slices.Values(j.joinColumnList))).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}

			if !yield(")", nil) {
				return
			}
		}
	}
}
