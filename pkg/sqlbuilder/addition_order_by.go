package sqlbuilder

import (
	"context"
)

func OrderBy(orders ...*Order) Addition {
	finalOrders := make([]*Order, 0)

	for i := range orders {
		if IsNilExpr(orders[i]) {
			continue
		}
		finalOrders = append(finalOrders, orders[i])
	}

	return &orderBy{
		orders: finalOrders,
	}
}

type orderBy struct {
	orders []*Order
}

func (orderBy) AdditionType() AdditionType {
	return AdditionOrderBy
}

func (o *orderBy) IsNil() bool {
	return o == nil || len(o.orders) == 0
}

func (o *orderBy) Ex(ctx context.Context) *Ex {
	e := Expr("ORDER BY ")
	for i := range o.orders {
		if i > 0 {
			e.WriteQueryByte(',')
		}
		e.WriteExpr(o.orders[i])
	}
	return e.Ex(ctx)
}

func AscOrder(target SqlExpr) *Order {
	return &Order{target: target, typ: "ASC"}
}

func DescOrder(target SqlExpr) *Order {
	return &Order{target: target, typ: "DESC"}
}

var _ SqlExpr = (*Order)(nil)

type Order struct {
	target SqlExpr
	typ    string
}

func (o *Order) IsNil() bool {
	return o == nil || IsNilExpr(o.target)
}

func (o *Order) Ex(ctx context.Context) *Ex {
	e := Expr("")
	e.Grow(1)

	e.WriteGroup(func(e *Ex) {
		e.WriteExpr(o.target)
	})

	if o.typ != "" {
		e.WriteQueryByte(' ')
		e.WriteQuery(o.typ)
	}

	return e.Ex(ctx)
}
