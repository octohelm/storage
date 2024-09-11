package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func OrderBy[M sqlbuilder.Model](orders ...*sqlbuilder.Order) TypedQuerierPatcher[M] {
	return &ordersPatcher[M]{orders: orders}
}

type ordersPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	orders []*sqlbuilder.Order
}

func (w ordersPatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	return q.OrderBy(w.orders...)
}
