package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func DistinctOn(on ...sqlfrag.Fragment) sqlfrag.Fragment {
	return &distinctOn{on: on}
}

type distinctOn struct {
	on []sqlfrag.Fragment
}

func (o *distinctOn) IsNil() bool {
	return o == nil || len(o.on) == 0
}

func (o *distinctOn) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("DISTINCT ON (", nil) {
			return
		}

		for g, args := range sqlfrag.Join(",", sqlfrag.NonNil(slices.Values(o.on))).Frag(ctx) {
			if !yield(g, args) {
				return
			}
		}

		if !yield(")", nil) {
			return
		}
	}
}

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

func NullsFirst() sqlfrag.Fragment {
	return sqlfrag.Pair(" NULLS FIRST")
}

func NullsLast() sqlfrag.Fragment {
	return sqlfrag.Pair(" NULLS LAST")
}

func DefaultOrder(target sqlfrag.Fragment, ex ...sqlfrag.Fragment) Order {
	return &order{target: target, ex: ex}
}

func AscOrder(target sqlfrag.Fragment, ex ...sqlfrag.Fragment) Order {
	return &order{target: target, typ: "ASC", ex: ex}
}

func DescOrder(target sqlfrag.Fragment, ex ...sqlfrag.Fragment) Order {
	return &order{target: target, typ: "DESC", ex: ex}
}

type Order interface {
	sqlfrag.Fragment

	orderType() string
}

type order struct {
	target sqlfrag.Fragment
	typ    string
	ex     []sqlfrag.Fragment
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

		if o.ex != nil {
			for _, x := range o.ex {
				if x.IsNil() {
					continue
				}
				for q, args := range x.Frag(ctx) {
					if !yield(q, args) {
						return
					}
				}
			}
		}
	}
}
