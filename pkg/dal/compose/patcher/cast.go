package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/x/slices"
)

func CastSliceAsAnySlice[M sqlbuilder.Model](patchers ...TypedQuerierPatcher[M]) []dal.QuerierPatcher {
	return slices.Map(patchers, func(e TypedQuerierPatcher[M]) dal.QuerierPatcher {
		return e
	})
}

func CastSlice[M sqlbuilder.Model](patchers ...dal.QuerierPatcher) []TypedQuerierPatcher[M] {
	return slices.Map(patchers, Cast[M])
}

func Cast[M sqlbuilder.Model](querierPatcher dal.QuerierPatcher) TypedQuerierPatcher[M] {
	return &patcher[M]{QuerierPatcher: querierPatcher}
}

type patcher[M sqlbuilder.Model] struct {
	dal.QuerierPatcher
}

func (patcher[M]) Model() *M {
	return new(M)
}
