package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

// EmptyCond 返回一个空条件。
func EmptyCond() SqlCondition {
	return (*Condition)(nil)
}

// SqlCondition 表示可组合的 SQL 条件。
type SqlCondition interface {
	sqlfrag.Fragment

	SqlConditionMarker
}

// SqlConditionMarker 用于标识条件类型。
type SqlConditionMarker interface {
	asCondition()
}

// AsCond 把片段包装为 Condition。
func AsCond(ex sqlfrag.Fragment) *Condition {
	if c, ok := ex.(*Condition); ok {
		return c
	}
	return &Condition{expr: ex}
}

// Condition 表示单个 SQL 条件片段。
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

// And 用 AND 组合多个条件。
func And(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("AND", progressCondition(conditions))
}

// AndSeq 用 AND 组合一个条件序列。
func AndSeq(conditions iter.Seq[sqlfrag.Fragment]) SqlCondition {
	return composedCondition("AND", progressCondition(slices.Collect(conditions)))
}

// Or 用 OR 组合多个条件。
func Or(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("OR", progressCondition(conditions))
}

// OrSeq 用 OR 组合一个条件序列。
func OrSeq(conditions iter.Seq[sqlfrag.Fragment]) SqlCondition {
	return composedCondition("OR", progressCondition(slices.Collect(conditions)))
}

// Xor 用 XOR 组合多个条件。
func Xor(conditions ...sqlfrag.Fragment) SqlCondition {
	return composedCondition("XOR", progressCondition(conditions))
}

// XorSeq 用 XOR 组合一个条件序列。
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

// ComposedCondition 表示由多个条件组合成的复合条件。
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
