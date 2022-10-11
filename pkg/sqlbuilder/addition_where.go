package sqlbuilder

import (
	"context"
)

func Where(c SqlExpr) Addition {
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
	return w == nil || IsNilExpr(w.condition)
}

func (w *where) Ex(ctx context.Context) *Ex {
	e := Expr("WHERE ")
	e.WriteExpr(w.condition)
	return e.Ex(ctx)
}
