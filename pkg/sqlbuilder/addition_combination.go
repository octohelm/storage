package sqlbuilder

import (
	"context"
)

type CombinationAddition interface {
	Addition
	All(stmtSelect SelectStatement) CombinationAddition
	Distinct(stmtSelect SelectStatement) CombinationAddition
}

func Union() CombinationAddition {
	return &combination{
		operator: "UNION",
	}
}

func Intersect() CombinationAddition {
	return &combination{
		operator: "INTERSECT",
	}
}

func Expect() CombinationAddition {
	return &combination{
		operator: "EXCEPT",
	}
}

type combination struct {
	operator   string // UNION | INTERSECT | EXCEPT
	method     string // ALL | DISTINCT
	stmtSelect SelectStatement
}

func (combination) AdditionType() AdditionType {
	return AdditionCombination
}

func (c *combination) IsNil() bool {
	return c == nil || IsNilExpr(c.stmtSelect)
}

func (c combination) All(stmtSelect SelectStatement) CombinationAddition {
	c.method = "ALL"
	c.stmtSelect = stmtSelect
	return &c
}

func (c combination) Distinct(stmtSelect SelectStatement) CombinationAddition {
	c.method = "DISTINCT"
	c.stmtSelect = stmtSelect
	return &c
}

func (c *combination) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(1)

	e.WriteQuery(c.operator)
	e.WriteQueryByte(' ')

	if c.method != "" {
		e.WriteQuery(c.method)
		e.WriteQueryByte(' ')
	}

	e.WriteExpr(c.stmtSelect)

	return e.Ex(ctx)
}
