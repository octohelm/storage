package session

import (
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

var catalogs = sync.Map{}

func RegisterCatalog(name string, catalog sqlbuilder.Catalog) {
	for t := range catalog.Tables() {
		catalogs.Store(t.TableName(), name)
	}
}
