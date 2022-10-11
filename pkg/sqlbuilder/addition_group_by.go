package sqlbuilder

import (
	"context"
)

type GroupByAddition interface {
	Addition
	Having(cond SqlExpr) GroupByAddition
}

func GroupBy(groups ...SqlExpr) GroupByAddition {
	return &groupBy{
		groups: groups,
	}
}

type groupBy struct {
	groups []SqlExpr
	// HAVING
	havingCond SqlExpr
}

func (groupBy) AdditionType() AdditionType {
	return AdditionGroupBy
}

func (g groupBy) Having(cond SqlExpr) GroupByAddition {
	g.havingCond = cond
	return &g
}

func (g *groupBy) IsNil() bool {
	return g == nil || len(g.groups) == 0
}

func (g *groupBy) Ex(ctx context.Context) *Ex {
	expr := Expr("GROUP BY ")

	for i, group := range g.groups {
		if i > 0 {
			expr.WriteQueryByte(',')
		}
		expr.WriteExpr(group)
	}

	if !(IsNilExpr(g.havingCond)) {
		expr.WriteQuery(" HAVING ")
		expr.WriteExpr(g.havingCond)
	}

	return expr.Ex(ctx)
}
