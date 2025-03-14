package querierpatcher

import (
	"github.com/octohelm/storage/deprecated/pkg/dal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type Typed[M sqlbuilder.Model] interface {
	dal.QuerierPatcher

	sqlbuilder.ModelNewer[M]
}
