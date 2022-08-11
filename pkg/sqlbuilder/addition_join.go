package sqlbuilder

import (
	"context"
	"strings"
)

type JoinAddition interface {
	Addition
	On(joinCondition SqlCondition) JoinAddition
	Using(joinColumnList ...Column) JoinAddition
}

func Join(table SqlExpr, prefixes ...string) JoinAddition {
	return &join{
		prefix: strings.Join(prefixes, " "),
		target: table,
	}
}

func InnerJoin(table SqlExpr) JoinAddition {
	return Join(table, "INNER")
}

func LeftJoin(table SqlExpr) JoinAddition {
	return Join(table, "LEFT")
}

func RightJoin(table SqlExpr) JoinAddition {
	return Join(table, "RIGHT")
}

func FullJoin(table SqlExpr) JoinAddition {
	return Join(table, "FULL")
}

func CrossJoin(table SqlExpr) JoinAddition {
	return Join(table, "CROSS")
}

type join struct {
	prefix         string
	target         SqlExpr
	joinCondition  SqlCondition
	joinColumnList []Column
}

func (j join) AdditionType() AdditionType {
	return AdditionJoin
}

func (j join) On(joinCondition SqlCondition) JoinAddition {
	j.joinCondition = joinCondition
	return &j
}

func (j join) Using(joinColumnList ...Column) JoinAddition {
	j.joinColumnList = joinColumnList
	return &j
}

func (j *join) IsNil() bool {
	return j == nil || IsNilExpr(j.target) || (j.prefix != "CROSS" && IsNilExpr(j.joinCondition) && len(j.joinColumnList) == 0)
}

func (j *join) Ex(ctx context.Context) *Ex {
	t := "JOIN "
	if j.prefix != "" {
		t = j.prefix + " " + t
	}

	e := Expr(t)

	e.WriteExpr(j.target)

	if !(IsNilExpr(j.joinCondition)) {
		e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
			ex := Expr(" ON ")
			ex.WriteExpr(j.joinCondition)
			return ex.Ex(ctx)
		}))
	}

	if len(j.joinColumnList) > 0 {
		e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
			ex := Expr(" USING ")

			ex.WriteGroup(func(e *Ex) {
				for i := range j.joinColumnList {
					if i != 0 {
						ex.WriteQuery(", ")
					}
					ex.WriteExpr(j.joinColumnList[i])
				}
			})

			return ex.Ex(ContextWithToggles(ctx, Toggles{
				ToggleMultiTable: false,
			}))
		}))
	}

	return e.Ex(ctx)
}
