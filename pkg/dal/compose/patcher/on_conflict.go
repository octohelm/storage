package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func OnConflictDoNothing[M sqlbuilder.Model](cols sqlbuilder.ColumnCollection) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols: cols,
	}
}

func OnConflictDoUpdateSet[M sqlbuilder.Model](cols sqlbuilder.ColumnCollection, toUpdates ...sqlbuilder.Column) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols:    cols,
		updates: toUpdates,
	}
}

func OnConflictDoWith[M sqlbuilder.Model](cols sqlbuilder.ColumnCollection, with func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols: cols,
		with: with,
	}
}

type onConflictPatcher[T any] struct {
	cols    sqlbuilder.ColumnCollection
	updates []sqlbuilder.Column
	with    func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition
}

func (o *onConflictPatcher[T]) ApplyMutation(m dal.Mutation[T]) dal.Mutation[T] {
	if o.with != nil {
		return m.OnConflict(o.cols).DoWith(o.with)
	}
	if len(o.updates) > 0 {
		return m.OnConflict(o.cols).DoUpdateSet(o.updates...)
	}
	return m.OnConflict(o.cols).DoNothing()
}
