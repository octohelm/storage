package session

import (
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

var catalogs = sync.Map{}

func RegisterCatalog(name string, tables *sqlbuilder.Tables) {
	tables.Range(func(tab sqlbuilder.Table, idx int) bool {
		catalogs.Store(tab.TableName(), name)
		return true
	})
}
