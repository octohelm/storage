package testutil

import "github.com/octohelm/storage/pkg/sqlbuilder"

func CatalogFrom(models ...sqlbuilder.Model) *sqlbuilder.Tables {
	tables := &sqlbuilder.Tables{}
	for i := range models {
		tables.Add(sqlbuilder.TableFromModel(models[i]))
	}
	return tables
}
