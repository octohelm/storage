package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
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

func (c *combination) IsNil() bool {
	return c == nil || sqlfrag.IsNil(c.stmtSelect)
}

func (c *combination) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield(c.operator+" ", nil) {
			return
		}

		if c.method != "" {
			if !yield(c.method+" ", nil) {
				return
			}
		}

		if !(yield(sqlfrag.All(ctx, c.stmtSelect))) {
			return
		}
	}
}
