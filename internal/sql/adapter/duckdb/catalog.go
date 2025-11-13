package duckdb

import (
	"bytes"
	"context"
	"strings"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func (a *duckdbAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	cat := &sqlbuilder.Tables{}

	tables := make([]duckdbTable, 0)

	{

		tTable := sqlbuilder.TableFromModel(&duckdbTable{})
		rows, err := a.Query(ctx, sqlbuilder.Select(sqlbuilder.ColumnCollect(tTable.Cols())).From(tTable))
		if err != nil {
			return nil, err
		}
		if err := scanner.Scan(ctx, rows, &tables); err != nil {
			return nil, err
		}
	}

	indexes := make([]duckdbIndex, 0)

	{
		tIndex := sqlbuilder.TableFromModel(&duckdbIndex{})
		rows, err := a.Query(ctx, sqlbuilder.Select(sqlbuilder.ColumnCollect(tIndex.Cols())).From(tIndex))
		if err != nil {
			return nil, err
		}
		if err := scanner.Scan(ctx, rows, &indexes); err != nil {
			return nil, err
		}

	}

	for _, schema := range tables {
		table := cat.Table(schema.Table)
		if table == nil {
			table = sqlbuilder.T(schema.Table)
			cat.Add(table)
		}

		fieldDefs := parseTableDecls(bytes.NewBufferString(schema.SQL))

		for fieldName, _ := range fieldDefs {
			if fieldName == "PRIMARY" {
				continue
			}

			def := sqlbuilder.ColumnDef{}

			//if pkSQL, ok := fieldDefs["PRIMARY"]; ok {
			//	def.AutoIncrement = strings.Contains(pkSQL, fieldName)
			//}

			table.(sqlbuilder.ColumnCollectionManger).AddCol(sqlbuilder.Col(fieldName, sqlbuilder.ColDef(def)))
		}
	}

	for _, schema := range indexes {
		if schema.SQL != "" {
			table := cat.Table(schema.Table)

			indexName := strings.ToLower(schema.Name[len(table.TableName())+1:])
			isUnique := strings.Contains(schema.SQL, "UNIQUE")
			indexColNameAndOptions := strings.Split(
				strings.TrimSpace(schema.SQL[strings.Index(schema.SQL, "(")+1:strings.Index(schema.SQL, ")")]),
				",",
			)

			colNameAndOptions := make([]sqlbuilder.FieldNameAndOption, len(indexColNameAndOptions))
			for i, c := range indexColNameAndOptions {
				colNameAndOptions[i] = sqlbuilder.FieldNameAndOption(strings.ReplaceAll(c, " ", ","))
			}

			var key sqlbuilder.Key

			if isUnique {
				key = sqlbuilder.UniqueIndex(indexName, nil, sqlbuilder.IndexFieldNameAndOptions(colNameAndOptions...))
			} else {
				key = sqlbuilder.Index(indexName, nil, sqlbuilder.IndexFieldNameAndOptions(colNameAndOptions...))
			}

			table.(sqlbuilder.KeyCollectionManager).AddKey(key)
		}
	}

	return cat, nil
}

type duckdbTable struct {
	Table string `db:"table_name"` // on <Table>
	SQL   string `db:"sql"`
}

func (duckdbTable) TableName() string {
	return "duckdb_tables"
}

type duckdbIndex struct {
	Table string `db:"table_name"`
	Name  string `db:"index_name"`
	SQL   string `db:"sql"`
}

func (duckdbIndex) TableName() string {
	return "duckdb_indexes"
}
