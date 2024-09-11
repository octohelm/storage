package patcher

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type TypedQuerierPatcher[M sqlbuilder.Model] interface {
	dal.QuerierPatcher

	sqlbuilder.ModelNewer[M]
}
