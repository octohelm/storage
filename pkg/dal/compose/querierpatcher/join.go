package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Join[M sqlbuilder.Model](table sqlbuilder.Table, on sqlbuilder.SqlExpr) Typed[M] {
	return &joinPatcher[M]{
		table: table,
		on:    on,
	}
}

type joinPatcher[M sqlbuilder.Model] struct {
	fromTable[M]
	table sqlbuilder.Table
	on    sqlbuilder.SqlExpr
}

func (w *joinPatcher[M]) Apply(q dal.Querier) dal.Querier {
	return q.Join(w.table, w.on)
}
