package duckdb

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"iter"
	"reflect"
	"slices"

	typex "github.com/octohelm/x/types"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

var _ adapter.Dialect = (*dialect)(nil)

type dialect struct{}

func (dialect) DriverName() string {
	return "duckdb"
}

func (c *dialect) AddIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		return nil
	}

	return sqlfrag.Pair("\nCREATE @index_type @index_name ON @table (@columnAndOptions);", sqlfrag.NamedArgSet{
		"table": sqlbuilder.GetKeyTable(key),
		"index_type": func() sqlfrag.Fragment {
			if key.IsUnique() {
				return sqlfrag.Const("UNIQUE INDEX")
			}
			return sqlfrag.Const("INDEX")
		}(),
		"index_name":       c.indexName(key),
		"columnAndOptions": sqlbuilder.AsKeyColumnsTableDef(key),
	})
}

func (c *dialect) DropIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		// pk could not changed
		return nil
	}

	return sqlfrag.Pair("\nDROP INDEX IF EXISTS @index;", sqlfrag.NamedArgSet{
		"index": c.indexName(key),
	})
}

func (c *dialect) CreateTableIsNotExists(t sqlbuilder.Table) (exprs []sqlfrag.Fragment) {

	exprs = append(exprs, sqlfrag.Pair("@seq\nCREATE TABLE IF NOT EXISTS @table (@def\n);", sqlfrag.NamedArgSet{
		"table": t,
		"seq": sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
			return func(yield func(string, []any) bool) {
				for col := range t.Cols() {
					def := sqlbuilder.GetColumnDef(col)
					if def.AutoIncrement {
						yield(fmt.Sprintf("\nCREATE SEQUENCE IF NOT EXISTS 'seq_%s' START 1;", t.TableName()), nil)
						return
					}
				}
			}
		}),
		"def": sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
			return func(yield func(string, []any) bool) {
				var autoIncrement sqlbuilder.Column

				idx := 0
				for col := range t.Cols() {
					def := sqlbuilder.GetColumnDef(col)

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

					for q, args := range c.DataType(def, t).Frag(ctx) {
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

	for _, key := range slices.SortedFunc(t.Keys(), func(a sqlbuilder.Key, b sqlbuilder.Key) int {
		return cmp.Compare(a.Name(), b.Name())
	}) {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
	}

	return exprs
}

func (c *dialect) DropTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Pair("\nDROP TABLE IF EXISTS @table;", sqlfrag.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) AddColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	t := sqlbuilder.GetColumnTable(col)

	return sqlfrag.Pair("\nALTER TABLE @table ADD COLUMN @col @dataType;", sqlfrag.NamedArgSet{
		"table":    t,
		"col":      col,
		"dataType": c.DataType(sqlbuilder.GetColumnDef(col), t),
	})
}

func (c *dialect) RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Pair("\nALTER TABLE @table RENAME COLUMN @oldCol TO @newCol;", sqlfrag.NamedArgSet{
		"table":  sqlbuilder.GetColumnTable(col),
		"oldCol": col,
		"newCol": target,
	})
}

func (c *dialect) ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlfrag.Fragment {
	colDef := sqlbuilder.GetColumnDef(col)
	prevDef := sqlbuilder.GetColumnDef(prev)

	if colDef.AutoIncrement {
		return nil
	}

	dbDataType := c.dataType(colDef.Type, colDef, sqlbuilder.GetColumnTable(col))
	prevDbDataType := c.dataType(prevDef.Type, prevDef, sqlbuilder.GetColumnTable(prev))

	actions := make([]sqlfrag.Fragment, 0)

	if dbDataType != prevDbDataType {
		actions = append(actions, sqlfrag.Pair(
			"ALTER COLUMN ? TYPE ? /* FROM ? */",
			col, sqlfrag.Const(dbDataType), sqlfrag.Const(prevDbDataType),
		))
	}

	if colDef.Null != prevDef.Null {
		action := "SET"
		if colDef.Null {
			action = "DROP"
		}

		actions = append(actions, sqlfrag.Pair(
			"ALTER COLUMN ? ? NOT NULL",
			col, sqlfrag.Const(action),
		))
	}

	defaultValue := normalizeDefaultValue(colDef.Default, dbDataType)
	prevDefaultValue := normalizeDefaultValue(prevDef.Default, prevDbDataType)

	if defaultValue != prevDefaultValue {
		if colDef.Default != nil {
			actions = append(actions, sqlfrag.Pair("ALTER COLUMN ? SET DEFAULT ? /* FROM ? */", col, sqlfrag.Const(defaultValue), sqlfrag.Const(prevDefaultValue)))
		} else {
			actions = append(actions, sqlfrag.Pair("ALTER COLUMN ? DROP DEFAULT", col))
		}
	}

	if len(actions) == 0 {
		return nil
	}

	return sqlfrag.Pair("\nALTER TABLE @table @actions;", sqlfrag.NamedArgSet{
		"table":   sqlbuilder.GetColumnTable(col),
		"actions": sqlfrag.JoinValues(", ", actions...),
	})
}

func (c *dialect) DropColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Pair("\nALTER TABLE @table DROP COLUMN @col;", sqlfrag.NamedArgSet{
		"table": sqlbuilder.GetColumnTable(col),
		"col":   col,
	})
}

func (c *dialect) DataType(columnType sqlbuilder.ColumnDef, t sqlbuilder.Table) sqlfrag.Fragment {
	dbDataType := c.dbDataType(columnType.Type, columnType, t)
	return sqlfrag.Pair(dbDataType + c.dataTypeModify(columnType, dbDataType))
}

func (c *dialect) dataType(typ typex.Type, columnType sqlbuilder.ColumnDef, t sqlbuilder.Table) string {
	return c.dbDataType(columnType.Type, columnType, t)
}

func (c *dialect) dbDataType(typ typex.Type, columnType sqlbuilder.ColumnDef, t sqlbuilder.Table) string {
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
		return fmt.Sprintf("INTEGER DEFAULT(nextval('seq_%s')) PRIMARY KEY", t.TableName())
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return c.dataType(typ.Elem(), columnType, t)
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

	if columnType.Default != nil {
		buf.WriteString(" DEFAULT(")
		buf.WriteString(normalizeDefaultValue(columnType.Default, dataType))
		buf.WriteString(")")
	}

	if !columnType.Null {
		if !columnType.AutoIncrement {
			buf.WriteString(" NOT NULL")
		}
	}

	return buf.String()
}

func (c dialect) indexName(key sqlbuilder.Key) sqlfrag.Fragment {
	return sqlfrag.Pair(fmt.Sprintf("%s_%s", sqlbuilder.GetKeyTable(key).TableName(), key.Name()))
}

func normalizeDefaultValue(defaultValue *string, dataType string) string {
	if defaultValue == nil {
		return ""
	}
	return *defaultValue
}
