package querierpatcher

import (
	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	slicesx "github.com/octohelm/x/slices"
)

func OrderBy[M sqlbuilder.Model](orders ...modelscoped.Order[M]) Typed[M] {
	return &ordersPatcher[M]{orders: orders}
}

type ordersPatcher[M sqlbuilder.Model] struct {
	modelscoped.M[M]

	orders []modelscoped.Order[M]
}

func (w *ordersPatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	return q.OrderBy(slicesx.Map(w.orders, func(e modelscoped.Order[M]) sqlbuilder.Order {
		return e
	})...)
}
