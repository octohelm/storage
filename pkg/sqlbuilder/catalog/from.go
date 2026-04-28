package catalog

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

// From 根据模型列表构造 catalog。
func From(models ...internal.Model) sqlbuilder.Catalog {
	tables := &sqlbuilder.Tables{}
	for i := range models {
		tables.Add(sqlbuilder.TableFromModel(models[i]))
	}
	return tables
}
