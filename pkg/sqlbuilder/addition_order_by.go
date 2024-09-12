package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func OrderBy(orders ...Order) Addition {
	finalOrders := make([]Order, 0)

	for i := range orders {
		if sqlfrag.IsNil(orders[i]) {
			continue
		}
		finalOrders = append(finalOrders, orders[i])
	}

	return &orderBy{
		orders: finalOrders,
	}
}

type orderBy struct {
	orders []Order
}

func (orderBy) AdditionType() AdditionType {
	return AdditionOrderBy
}

func (o *orderBy) IsNil() bool {
	return o == nil || len(o.orders) == 0
}

func (o *orderBy) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("ORDER BY ", nil) {
			return
		}

		for g, args := range sqlfrag.Join(",", sqlfrag.NonNil(slices.Values(o.orders))).Frag(ctx) {
			if !yield(g, args) {
				return
			}
		}
	}
}

func AscOrder(target sqlfrag.Fragment) Order {
	return &order{target: target, typ: "ASC"}
}

func DescOrder(target sqlfrag.Fragment) Order {
	return &order{target: target, typ: "DESC"}
}

type Order interface {
	sqlfrag.Fragment

	orderType() string
}

type order struct {
	target sqlfrag.Fragment
	typ    string
}

func (o *order) orderType() string {
	return o.typ
}

func (o *order) IsNil() bool {
	return o == nil || sqlfrag.IsNil(o.target)
}

func (o *order) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		for q, args := range sqlfrag.InlineBlock(o.target).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}

		if o.typ != "" {
			if !yield(" "+o.typ, nil) {
				return
			}
		}
	}
}
