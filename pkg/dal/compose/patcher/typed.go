package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type TypedQuerierPatcher[M sqlbuilder.Model] interface {
	dal.QuerierPatcher

	Model() *M
}

type fromTable[M sqlbuilder.Model] struct{}

func (fromTable[M]) Model() *M {
	return new(M)
}
