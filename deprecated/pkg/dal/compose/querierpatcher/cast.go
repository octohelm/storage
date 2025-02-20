package querierpatcher

import (
	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/x/slices"
)

func CastSliceAsAnySlice[M sqlbuilder.Model](patchers ...Typed[M]) []dal.QuerierPatcher {
	return slices.Map(patchers, func(e Typed[M]) dal.QuerierPatcher {
		return e
	})
}

func CastSlice[M sqlbuilder.Model](patchers ...dal.QuerierPatcher) []Typed[M] {
	return slices.Map(patchers, Cast[M])
}

func Cast[M sqlbuilder.Model](querierPatcher dal.QuerierPatcher) Typed[M] {
	return &patcher[M]{QuerierPatcher: querierPatcher}
}

type patcher[M sqlbuilder.Model] struct {
	dal.QuerierPatcher
}

func (patcher[M]) Model() *M {
	return new(M)
}
