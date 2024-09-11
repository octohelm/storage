package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	slicesx "github.com/octohelm/x/slices"
)

func OnConflictDoNothing[M sqlbuilder.Model](cols modelscoped.ColumnSeq[M]) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols: cols,
	}
}

func OnConflictDoUpdateSet[M sqlbuilder.Model](cols modelscoped.ColumnSeq[M], toUpdates ...modelscoped.Column[M]) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols:    cols,
		updates: toUpdates,
	}
}

func OnConflictDoWith[M sqlbuilder.Model](cols modelscoped.ColumnSeq[M], with func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition) dal.MutationPatcher[M] {
	return &onConflictPatcher[M]{
		cols: cols,
		with: with,
	}
}

type onConflictPatcher[M sqlbuilder.Model] struct {
	cols    modelscoped.ColumnSeq[M]
	updates []modelscoped.Column[M]
	with    func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition
}

func (o *onConflictPatcher[M]) ApplyMutation(m dal.Mutation[M]) dal.Mutation[M] {
	if o.with != nil {
		return m.OnConflict(o.cols).DoWith(o.with)
	}
	if len(o.updates) > 0 {
		return m.OnConflict(o.cols).DoUpdateSet(
			slicesx.Map(o.updates, func(e modelscoped.Column[M]) sqlbuilder.Column {
				return e
			})...,
		)
	}
	return m.OnConflict(o.cols).DoNothing()
}
