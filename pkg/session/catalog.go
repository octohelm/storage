package session

import (
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

var catalogs = sync.Map{}

// RegisterCatalog 按表名注册 catalog 对应的会话名。
func RegisterCatalog(name string, catalog sqlbuilder.Catalog) {
	for t := range catalog.Tables() {
		catalogs.Store(t.TableName(), name)
	}
}
