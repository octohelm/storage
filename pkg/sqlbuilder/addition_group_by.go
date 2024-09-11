package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type GroupByAddition interface {
	Addition

	Having(cond sqlfrag.Fragment) GroupByAddition
}

func GroupBy(groups ...sqlfrag.Fragment) GroupByAddition {
	return &groupBy{
		groups: groups,
	}
}

type groupBy struct {
	groups []sqlfrag.Fragment
	// HAVING
	havingCond sqlfrag.Fragment
}

func (groupBy) AdditionType() AdditionType {
	return AdditionGroupBy
}

func (g groupBy) Having(cond sqlfrag.Fragment) GroupByAddition {
	g.havingCond = cond
	return &g
}

func (g *groupBy) IsNil() bool {
	return g == nil || len(g.groups) == 0
}

func (g *groupBy) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("GROUP BY ", nil) {
			return
		}

		for i, group := range g.groups {
			if i > 0 {
				if !yield(",", nil) {
					return
				}
			}

			if !yield(sqlfrag.All(ctx, group)) {
				return
			}
		}

		if !(sqlfrag.IsNil(g.havingCond)) {
			if !yield(" HAVING ", nil) {
				return
			}

			if !yield(sqlfrag.All(ctx, g.havingCond)) {
				return
			}
		}
	}
}
