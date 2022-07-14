package postgres

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	textscanner "text/scanner"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func (a *pgAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	return catalog(ctx, a, a.dbName)
}

var reUsing = regexp.MustCompile(`USING ([^ ]+)`)

func catalog(ctx context.Context, a adapter.Adapter, dbName string) (*sqlbuilder.Tables, error) {
	cat := &sqlbuilder.Tables{}

	tableColumnSchema := sqlbuilder.TableFromModel(&columnSchema{})

	tableSchema := "public"

	stmt := sqlbuilder.Select(tableColumnSchema.Cols()).From(tableColumnSchema,
		sqlbuilder.Where(
			sqlbuilder.And(
				tableColumnSchema.F("TABLE_SCHEMA").Eq(tableSchema),
			),
		),
	)

	rows, err := a.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	colSchemaList := make([]columnSchema, 0)

	if err := scanner.Scan(ctx, rows, &colSchemaList); err != nil {
		return nil, err
	}

	for i := range colSchemaList {
		colSchema := colSchemaList[i]

		table := cat.Table(colSchema.TABLE_NAME)
		if table == nil {
			table = sqlbuilder.T(colSchema.TABLE_NAME)
			cat.Add(table)
		}

		table.(sqlbuilder.ColumnCollectionManger).AddCol(colSchema.ToColumn())
	}

	if cols := tableColumnSchema.Cols(); cols.Len() != 0 {
		tableIndexSchema := sqlbuilder.TableFromModel(&indexSchema{})

		indexList := make([]indexSchema, 0)

		rows, err := a.Query(
			ctx,
			sqlbuilder.Select(tableIndexSchema.Cols()).
				From(
					tableIndexSchema,
					sqlbuilder.Where(
						sqlbuilder.And(
							tableIndexSchema.F("TABLE_SCHEMA").Eq(tableSchema),
						),
					),
				),
		)
		if err != nil {
			return nil, err
		}

		if err := scanner.Scan(ctx, rows, &indexList); err != nil {
			return nil, err
		}

		for _, idxSchema := range indexList {
			t := cat.Table(idxSchema.TABLE_NAME)
			t.(sqlbuilder.KeyCollectionManager).AddKey(idxSchema.ToKey(t))
		}
	}

	return cat, nil
}

type columnSchema struct {
	TABLE_SCHEMA             string `db:"table_schema"`
	TABLE_NAME               string `db:"table_name"`
	COLUMN_NAME              string `db:"column_name"`
	DATA_TYPE                string `db:"data_type"`
	IS_NULLABLE              string `db:"is_nullable"`
	COLUMN_DEFAULT           string `db:"column_default"`
	CHARACTER_MAXIMUM_LENGTH uint64 `db:"character_maximum_length"`
	NUMERIC_PRECISION        uint64 `db:"numeric_precision"`
	NUMERIC_SCALE            uint64 `db:"numeric_scale"`
}

func (columnSchema) TableName() string {
	return "information_schema.columns"
}

func (columnSchema *columnSchema) ToColumn() sqlbuilder.Column {
	defaultValue := columnSchema.COLUMN_DEFAULT
	def := sqlbuilder.ColumnDef{}

	if defaultValue != "" {
		def.AutoIncrement = strings.HasSuffix(columnSchema.COLUMN_DEFAULT, "_seq'::regclass)")

		if !def.AutoIncrement {
			if !strings.Contains(defaultValue, "'::") && '0' <= defaultValue[0] && defaultValue[0] <= '9' {
				defaultValue = fmt.Sprintf("'%s'::integer", defaultValue)
			}
			def.Default = &defaultValue
		}
	}

	dataType := columnSchema.DATA_TYPE

	if def.AutoIncrement {
		if strings.HasPrefix(dataType, "big") {
			dataType = "bigserial"
		} else {
			dataType = "serial"
		}
	}

	def.DataType = dataType

	// numeric type
	if columnSchema.NUMERIC_PRECISION > 0 {
		def.Length = columnSchema.NUMERIC_PRECISION
		def.Decimal = columnSchema.NUMERIC_SCALE
	} else {
		def.Length = columnSchema.CHARACTER_MAXIMUM_LENGTH
	}

	if columnSchema.IS_NULLABLE == "YES" {
		def.Null = true
	}

	return sqlbuilder.Col(columnSchema.COLUMN_NAME, sqlbuilder.ColDef(def))
}

type indexSchema struct {
	TABLE_SCHEMA string `db:"schemaname"`
	TABLE_NAME   string `db:"tablename"`
	INDEX_NAME   string `db:"indexname"`
	INDEX_DEF    string `db:"indexdef"`
}

func (indexSchema) TableName() string {
	return "pg_indexes"
}

func (idxSchema *indexSchema) ToKey(table sqlbuilder.Table) sqlbuilder.Key {
	name := strings.ToLower(idxSchema.INDEX_NAME[len(table.TableName())+1:])
	method := strings.ToUpper(reUsing.FindString(idxSchema.INDEX_DEF)[6:])
	isUnique := strings.Contains(idxSchema.INDEX_DEF, "UNIQUE")

	colNameAndOptions := make([]string, 0)

	s := &textscanner.Scanner{}
	s.Init(bytes.NewBufferString(strings.TrimSpace(reUsing.Split(idxSchema.INDEX_DEF, 2)[1])))

	parts := make([]string, 0)

	for t := s.Scan(); t != textscanner.EOF; t = s.Scan() {
		part := s.TokenText()
		switch part {
		case "(":
			continue
		case ",", ")":
			colNameAndOption := parts[0]
			if len(parts) > 1 {
				colNameAndOption += "/" + parts[1]
			}
			colNameAndOptions = append(colNameAndOptions, colNameAndOption)
			continue
		}
		parts = append(parts, part)
	}

	if isUnique {
		return sqlbuilder.UniqueIndex(name, nil, sqlbuilder.IndexUsing(method), sqlbuilder.IndexColNameAndOptions(colNameAndOptions...))
	}
	return sqlbuilder.Index(name, nil, sqlbuilder.IndexUsing(method), sqlbuilder.IndexColNameAndOptions(colNameAndOptions...))
}
