package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Alias(expr sqlfrag.Fragment, name string) sqlfrag.Fragment {
	return &exAlias{
		name:     name,
		Fragment: expr,
	}
}

type exAlias struct {
	name string

	sqlfrag.Fragment
}

func (alias *exAlias) IsNil() bool {
	return alias == nil || alias.name == "" || sqlfrag.IsNil(alias.Fragment)
}

func (alias *exAlias) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.Pair("? AS ?", alias.Fragment, sqlfrag.Const(alias.name)).Frag(ContextWithToggles(ctx, Toggles{
		ToggleNeedAutoAlias: false,
	}))
}

func MultiMayAutoAlias(columns ...sqlfrag.Fragment) sqlfrag.Fragment {
	return &exMayAutoAlias{
		columns: slices.Collect(sqlfrag.NonNil(slices.Values(columns))),
	}
}

type exMayAutoAlias struct {
	columns []sqlfrag.Fragment
}

func (alias *exMayAutoAlias) IsNil() bool {
	return alias == nil || len(alias.columns) == 0
}

func (alias *exMayAutoAlias) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		ctx = ContextWithToggles(ctx, Toggles{
			ToggleNeedAutoAlias: true,
		})

		for q, args := range sqlfrag.JoinValues(", ", alias.columns...).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
