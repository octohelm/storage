package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func OrderBy[M sqlbuilder.Model](orders ...*sqlbuilder.Order) Typed[M] {
	return &ordersPatcher[M]{orders: orders}
}

type ordersPatcher[M sqlbuilder.Model] struct {
	fromTable[M]

	orders []*sqlbuilder.Order
}

func (w ordersPatcher[M]) Apply(q dal.Querier) dal.Querier {
	return q.OrderBy(w.orders...)
}
