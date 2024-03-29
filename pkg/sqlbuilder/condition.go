package sqlbuilder

import (
	"context"
)

func EmptyCond() SqlCondition {
	return (*Condition)(nil)
}

type SqlCondition interface {
	SqlExpr
	SqlConditionMarker
}

type SqlConditionMarker interface {
	asCondition()
}

func AsCond(ex SqlExpr) *Condition {
	if c, ok := ex.(*Condition); ok {
		return c
	}
	return &Condition{expr: ex}
}

type Condition struct {
	expr SqlExpr
	SqlConditionMarker
}

func (c *Condition) Ex(ctx context.Context) *Ex {
	if IsNilExpr(c.expr) {
		return nil
	}
	return c.expr.Ex(ctx)
}

func (c *Condition) IsNil() bool {
	return c == nil || IsNilExpr(c.expr)
}

func And(conditions ...SqlExpr) SqlCondition {
	return composedCondition("AND", filterNilCondition(conditions)...)
}

func Or(conditions ...SqlExpr) SqlCondition {
	return composedCondition("OR", filterNilCondition(conditions)...)
}

func Xor(conditions ...SqlExpr) SqlCondition {
	return composedCondition("XOR", filterNilCondition(conditions)...)
}

func filterNilCondition(conditions []SqlExpr) []SqlCondition {
	finals := make([]SqlCondition, 0, len(conditions))

	for i := range conditions {
		condition := AsCond(conditions[i])
		if IsNilExpr(condition) {
			continue
		}
		finals = append(finals, condition)
	}

	return finals
}

func composedCondition(op string, conditions ...SqlCondition) SqlCondition {
	return &ComposedCondition{op: op, conditions: conditions}
}

type ComposedCondition struct {
	SqlConditionMarker

	op         string
	conditions []SqlCondition
}

func (c *ComposedCondition) And(cond SqlCondition) SqlCondition {
	return And(c, cond)
}

func (c *ComposedCondition) Or(cond SqlCondition) SqlCondition {
	return Or(c, cond)
}

func (c *ComposedCondition) Xor(cond SqlCondition) SqlCondition {
	return Xor(c, cond)
}

func (c *ComposedCondition) IsNil() bool {
	if c == nil {
		return true
	}
	if c.op == "" {
		return true
	}

	isNil := true

	for i := range c.conditions {
		if !IsNilExpr(c.conditions[i]) {
			isNil = false
			continue
		}
	}

	return isNil
}

func (c *ComposedCondition) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(len(c.conditions))

	for i := range c.conditions {
		condition := c.conditions[i]

		if i > 0 {
			e.WriteQueryByte(' ')
			e.WriteQuery(c.op)
			e.WriteQueryByte(' ')
		}

		e.WriteGroup(func(e *Ex) {
			e.WriteExpr(condition)
		})
	}

	return e.Ex(ctx)
}
