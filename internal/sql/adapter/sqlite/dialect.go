package sqlite

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"

	typex "github.com/octohelm/x/types"
)

var _ adapter.Dialect = (*dialect)(nil)

type dialect struct{}

func (dialect) DriverName() string {
	return "sqlite"
}

func (c *dialect) AddIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		return sqlfrag.Pair("ALTER TABLE ? ADD PRIMARY KEY (?);", key.T(), sqlbuilder.ColumnCollect(key.Cols()))
	}

	return sqlfrag.Pair("CREATE @index_type @index_name ON @table (@columns);", sqlfrag.NamedArgSet{
		"table": key.T(),
		"index_type": func() sqlfrag.Fragment {
			if key.IsUnique() {
				return sqlfrag.Const("UNIQUE INDEX")
			}
			return sqlfrag.Const("INDEX")
		}(),
		"index_name": c.indexName(key),
		"columns":    sqlbuilder.ColumnCollect(key.Cols()),
	})
}

func (c *dialect) DropIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		// pk could not changed
		return nil
	}

	return sqlfrag.Pair("DROP INDEX IF EXISTS @index;", sqlfrag.NamedArgSet{
		"index": c.indexName(key),
	})
}

func (c *dialect) CreateTableIsNotExists(t sqlbuilder.Table) (exprs []sqlfrag.Fragment) {
	exprs = append(exprs, sqlfrag.Pair("CREATE TABLE IF NOT EXISTS @table (@def\n);", sqlfrag.NamedArgSet{
		"table": t,
		"def": sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
			return func(yield func(string, []any) bool) {
				var autoIncrement sqlbuilder.Column

				idx := 0
				for col := range t.Cols() {
					def := col.Def()

					// skip deprecated col
					if def.DeprecatedActions != nil {
						continue
					}

					if def.AutoIncrement {
						autoIncrement = col
					}

					if idx > 0 {
						if !yield(",", nil) {
							return
						}
					}
					idx++

					if !yield("\n\t", nil) {
						return
					}

					for q, args := range col.Frag(ctx) {
						if !yield(q, args) {
							return
						}
					}

					if !yield(" ", nil) {
						return
					}

					for q, args := range c.DataType(def).Frag(ctx) {
						if !yield(q, args) {
							return
						}
					}
				}

				for key := range t.Keys() {
					if key.IsPrimary() {
						skip := false

						if autoIncrement != nil {
							for col := range key.Cols() {
								if autoIncrement.Name() == col.Name() {
									skip = true
									// auto increment pk will create when table define
									break
								}
							}
						}

						if skip {
							continue
						}

						for q, args := range sqlfrag.Pair(",\n\tPRIMARY KEY (?)", sqlbuilder.ColumnCollect(key.Cols())).Frag(ctx) {
							if !yield(q, args) {
								return
							}
						}
					}
				}
			}
		}),
	}))

	for key := range t.Keys() {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
	}

	return
}

func (c *dialect) DropTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Pair("DROP TABLE IF EXISTS @table;", sqlfrag.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) TruncateTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Pair("TRUNCATE TABLE @table;", sqlfrag.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) AddColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Pair("ALTER TABLE @table ADD COLUMN @col @dataType;", sqlfrag.NamedArgSet{
		"table":    col.T(),
		"col":      col,
		"dataType": c.DataType(col.Def()),
	})
}

func (c *dialect) RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Pair("ALTER TABLE @table RENAME COLUMN @oldCol TO @newCol;", sqlfrag.NamedArgSet{
		"table":  col.T(),
		"oldCol": col,
		"newCol": target,
	})
}

func (c *dialect) ModifyColumn(col sqlbuilder.Column, prevCol sqlbuilder.Column) sqlfrag.Fragment {
	def := col.Def()

	// incr id never modified
	if def.AutoIncrement {
		return nil
	}

	prevTmpCol := sqlbuilder.Col("__"+prevCol.Name(), sqlbuilder.ColDef(prevCol.Def())).Of(prevCol.T())

	return sqlfrag.JoinValues("",
		sqlfrag.Pair("ALTER TABLE @table RENAME COLUMN @prevCol TO @tmpCol;", sqlfrag.NamedArgSet{
			"table":   prevCol.T(),
			"prevCol": prevCol,
			"tmpCol":  prevTmpCol,
		}),

		c.AddColumn(col),

		sqlfrag.Pair("UPDATE @table SET @col = @tmpCol;", sqlfrag.NamedArgSet{
			"table":  col.T(),
			"col":    col,
			"tmpCol": prevTmpCol,
		}),

		c.DropColumn(prevTmpCol),
	)
}

func (c *dialect) DropColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Pair("ALTER TABLE @table DROP COLUMN @col;", sqlfrag.NamedArgSet{
		"table": col.T(),
		"col":   col,
	})
}

func (c *dialect) DataType(columnType sqlbuilder.ColumnDef) sqlfrag.Fragment {
	dbDataType := c.dbDataType(columnType.Type, columnType)
	return sqlfrag.Pair(dbDataType + c.dataTypeModify(columnType, dbDataType))
}

func (c *dialect) dataType(typ typex.Type, columnType sqlbuilder.ColumnDef) string {
	return c.dbDataType(columnType.Type, columnType)
}

func (c *dialect) dbDataType(typ typex.Type, columnType sqlbuilder.ColumnDef) string {
	if columnType.DataType != "" {
		// for type from catalog
		return columnType.DataType
	}

	if rv, ok := typex.TryNew(typ); ok {
		v := rv.Interface()

		if dtd, ok := v.(sqlbuilder.DataTypeDescriber); ok {
			return dtd.DataType(c.DriverName())
		}
	}

	if columnType.AutoIncrement {
		return "INTEGER PRIMARY KEY AUTOINCREMENT"
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return c.dataType(typ.Elem(), columnType)
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "INTEGER"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Uint64:
		return "UNSIGNED BIG INT"
	case reflect.Float32:
		return "FLOAT"
	case reflect.Float64:
		return "DOUBLE"
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "BLOB"
		}
	case reflect.String:
		return "TEXT"
	default:
		if typ.Name() == "Time" && typ.PkgPath() == "time" {
			return "DATETIME"
		}
	}

	panic(fmt.Errorf("unsupport type %s", typ))
}

func (c *dialect) dataTypeModify(columnType sqlbuilder.ColumnDef, dataType string) string {
	buf := bytes.NewBuffer(nil)

	if !columnType.Null {
		buf.WriteString(" NOT NULL")
	}

	if columnType.Default != nil {
		buf.WriteString(" DEFAULT ")
		buf.WriteString(normalizeDefaultValue(columnType.Default, dataType))
	}

	return buf.String()
}

func (c dialect) indexName(key sqlbuilder.Key) sqlfrag.Fragment {
	return sqlfrag.Pair(fmt.Sprintf("%s_%s", key.T().TableName(), key.Name()))
}

func normalizeDefaultValue(defaultValue *string, dataType string) string {
	if defaultValue == nil {
		return ""
	}
	return *defaultValue
}
