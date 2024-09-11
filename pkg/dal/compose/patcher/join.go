package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Join[M sqlbuilder.Model](table sqlbuilder.Table, on sqlbuilder.SqlExpr) TypedQuerierPatcher[M] {
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

func (w *joinPatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	return q.Join(w.table, w.on)
}
