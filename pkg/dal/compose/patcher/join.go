package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Join[M sqlbuilder.Model](table sqlbuilder.Table, on sqlfrag.Fragment) TypedQuerierPatcher[M] {
	return &joinPatcher[M]{
		table: table,
		on:    on,
	}
}

type joinPatcher[M sqlbuilder.Model] struct {
	modelscoped.M[M]

	table sqlbuilder.Table
	on    sqlfrag.Fragment
}

func (w *joinPatcher[M]) ApplyQuerier(q dal.Querier) dal.Querier {
	return q.Join(w.table, w.on)
}
