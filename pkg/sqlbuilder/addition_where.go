package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Where(c sqlfrag.Fragment) Addition {
	switch x := c.(type) {
	case *where:
		return x
	}

	return &where{
		condition: AsCond(c),
	}
}

type where struct {
	condition SqlCondition
}

func (*where) AdditionType() AdditionType {
	return AdditionWhere
}

func (w *where) IsNil() bool {
	return w == nil || sqlfrag.IsNil(w.condition)
}

func (w *where) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("WHERE ", nil) {
			return
		}

		for q, args := range w.condition.Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
