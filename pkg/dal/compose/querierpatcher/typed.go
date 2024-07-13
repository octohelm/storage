package querierpatcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type Typed[M sqlbuilder.Model] interface {
	Model() *M

	dal.QuerierPatcher
}

type fromTable[M sqlbuilder.Model] struct{}

func (fromTable[M]) Model() *M {
	return new(M)
}
