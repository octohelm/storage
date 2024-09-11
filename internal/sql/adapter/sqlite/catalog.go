package sqlite

import (
	"bytes"
	"context"
	"io"
	"strings"
	textscanner "text/scanner"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func (a *sqliteAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	cat := &sqlbuilder.Tables{}

	tblSqlMaster := sqlbuilder.TableFromModel(&sqliteMaster{})

	stmt := sqlbuilder.Select(sqlbuilder.ColumnCollect(tblSqlMaster.Cols())).From(tblSqlMaster)

	rows, err := a.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	schemaList := make([]sqliteMaster, 0)

	if err := scanner.Scan(ctx, rows, &schemaList); err != nil {
		return nil, err
	}

	for _, schema := range schemaList {
		if schema.Type == "table" {
			table := cat.Table(schema.Table)
			if table == nil {
				table = sqlbuilder.T(schema.Table)
				cat.Add(table)
			}

			cols := extractCols(bytes.NewBufferString(schema.SQL))
			for f, colSql := range cols {
				if f == "PRIMARY" {
					continue
				}

				def := sqlbuilder.ColumnDef{}

				if pkSQL, ok := cols["PRIMARY"]; ok {
					def.AutoIncrement = strings.Contains(pkSQL, f)
				}

				defaultValue := ""
				parts := strings.Split(colSql, " DEFAULT ")
				if len(parts) > 1 {
					defaultValue = parts[1]
					def.Default = &defaultValue
				}

				def.Null = !strings.Contains(parts[0], "NOT NULL")

				if !def.Null {
					def.DataType = strings.TrimSpace(strings.Split(parts[0], "NOT NULL")[0])
				} else {
					def.DataType = strings.TrimSpace(strings.Split(parts[0], "NULL")[0])
				}

				table.(sqlbuilder.ColumnCollectionManger).AddCol(sqlbuilder.Col(f, sqlbuilder.ColDef(def)))
			}

		}
	}

	for _, schema := range schemaList {
		if schema.Type == "index" && schema.SQL != "" {
			table := cat.Table(schema.Table)

			indexName := strings.ToLower(schema.Name[len(table.TableName())+1:])
			isUnique := strings.Contains(schema.SQL, "UNIQUE")
			indexColNameAndOptions := strings.Split(
				strings.TrimSpace(schema.SQL[strings.Index(schema.SQL, "(")+1:strings.Index(schema.SQL, ")")]),
				",",
			)

			var key sqlbuilder.Key

			if isUnique {
				key = sqlbuilder.UniqueIndex(indexName, nil, sqlbuilder.IndexColNameAndOptions(indexColNameAndOptions...))
			} else {
				key = sqlbuilder.Index(indexName, nil, sqlbuilder.IndexColNameAndOptions(indexColNameAndOptions...))
			}

			table.(sqlbuilder.KeyCollectionManager).AddKey(key)
		}
	}

	return cat, nil
}

type sqliteMaster struct {
	Type  string `db:"type"` // index or table
	Name  string `db:"name"`
	SQL   string `db:"sql"`
	Table string `db:"tbl_name"` // on <Table>
}

func (sqliteMaster) TableName() string {
	return "sqlite_master"
}

func extractCols(r io.Reader) map[string]string {
	s := &textscanner.Scanner{}
	s.Init(r)
	s.Error = func(s *textscanner.Scanner, msg string) {}

	scope := 0
	cols := make(map[string]string)
	parts := make([]string, 0)

	collect := func() {
		if len(parts) == 0 || scope != 1 {
			return
		}
		cols[parts[0]] = strings.Join(parts[1:], " ")
		parts = make([]string, 0)
	}

	for tok := s.Scan(); tok != textscanner.EOF; tok = s.Scan() {
		part := s.TokenText()

		switch part {
		case "(":
			scope++
			if scope == 1 {
				continue
			}
		case ")":
			collect()
			scope--
		case ",":
			collect()
			continue
		}

		if scope > 0 {
			parts = append(parts, part)
		}
	}

	return cols
}
