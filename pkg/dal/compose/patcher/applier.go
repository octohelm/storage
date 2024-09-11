package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func ApplyToQuerier[M sqlbuilder.Model](q dal.Querier, patchers ...TypedQuerierPatcher[M]) dal.Querier {
	return q.Apply(CastSliceAsAnySlice(patchers...)...)
}

func ApplyToMutation[M sqlbuilder.Model](q dal.Mutation[M], patchers ...dal.MutationPatcher[M]) dal.Mutation[M] {
	return q.Apply(patchers...)
}
