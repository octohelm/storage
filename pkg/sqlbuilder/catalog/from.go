package catalog

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

func From(models ...internal.Model) *sqlbuilder.Tables {
	tables := &sqlbuilder.Tables{}
	for i := range models {
		tables.Add(sqlbuilder.TableFromModel(models[i]))
	}
	return tables
}
