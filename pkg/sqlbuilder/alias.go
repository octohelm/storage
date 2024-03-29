package sqlbuilder

import (
	"context"
)

func Alias(expr SqlExpr, name string) SqlExpr {
	return &exAlias{
		name:    name,
		SqlExpr: expr,
	}
}

type exAlias struct {
	name string
	SqlExpr
}

func (alias *exAlias) IsNil() bool {
	return alias == nil || alias.name == "" || IsNilExpr(alias.SqlExpr)
}

func (alias *exAlias) Ex(ctx context.Context) *Ex {
	return Expr("? AS ?", alias.SqlExpr, Expr(alias.name)).Ex(ContextWithToggles(ctx, Toggles{
		ToggleNeedAutoAlias: false,
	}))
}

func MultiMayAutoAlias(columns ...SqlExpr) SqlExpr {
	return &exMayAutoAlias{
		columns: columns,
	}
}

type exMayAutoAlias struct {
	columns []SqlExpr
}

func (alias *exMayAutoAlias) IsNil() bool {
	return alias == nil || len(alias.columns) == 0
}

func (alias *exMayAutoAlias) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(1)

	RangeNotNilExpr(alias.columns, func(expr SqlExpr, i int) {
		if i > 0 {
			e.WriteQuery(", ")
		}
		e.WriteExpr(expr)
	})

	return e.Ex(ContextWithToggles(ctx, Toggles{
		ToggleNeedAutoAlias: true,
	}))
}
