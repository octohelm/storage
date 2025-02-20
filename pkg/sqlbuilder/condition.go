package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func EmptyCond() SqlCondition {
	return (*Condition)(nil)
}

type SqlCondition interface {
	sqlfrag.Fragment

	SqlConditionMarker
}

type SqlConditionMarker interface {
	asCondition()
}

func AsCond(ex sqlfrag.Fragment) *Condition {
	if c, ok := ex.(*Condition); ok {
		return c
	}
	return &Condition{expr: ex}
}

type Condition struct {
	expr sqlfrag.Fragment

	SqlConditionMarker
}

func (c *Condition) Frag(ctx context.Context) iter.Seq2[string, []any] {
	if sqlfrag.IsNil(c.expr) {
		return nil
	}

	return c.expr.Frag(ctx)
}

func (c *Condition) IsNil() bool {
	return c == nil || sqlfrag.IsNil(c.expr)
}

func And(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("AND", progressCondition(conditions))
}

func AndSeq(conditions iter.Seq[sqlfrag.Fragment]) SqlCondition {
	return composedCondition("AND", progressCondition(slices.Collect(conditions)))
}

func Or(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("OR", progressCondition(conditions))
}

func OrSeq(conditions iter.Seq[sqlfrag.Fragment]) SqlCondition {
	return composedCondition("OR", progressCondition(slices.Collect(conditions)))
}

func Xor(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("XOR", progressCondition(conditions))
}

func XorSeq(conditions iter.Seq[sqlfrag.Fragment]) SqlCondition {
	return composedCondition("XOR", progressCondition(slices.Collect(conditions)))
}

func progressCondition(conditions []sqlfrag.Fragment) []SqlCondition {
	finals := make([]SqlCondition, 0, len(conditions))

	for i := range conditions {
		c := conditions[i]

		switch x := conditions[i].(type) {
		case *where:
			c = x.condition
		default:

		}

		if sqlfrag.IsNil(c) {
			continue
		}

		finals = append(finals, AsCond(c))
	}

	return finals
}

func composedCondition(op string, conditions []SqlCondition) SqlCondition {
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
		if !sqlfrag.IsNil(c.conditions[i]) {
			isNil = false
			continue
		}
	}

	return isNil
}

func (c *ComposedCondition) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if len(c.conditions) == 1 {
			for q, args := range c.conditions[0].Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
			return
		}

		for i, cond := range c.conditions {
			if i > 0 {
				if !yield(" "+c.op+" ", nil) {
					return
				}
			}

			for q, args := range sqlfrag.InlineBlock(cond).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
		}
	}
}
