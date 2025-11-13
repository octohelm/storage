package postgres

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"reflect"
	"strconv"
	"strings"

	typex "github.com/octohelm/x/types"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

var _ adapter.Dialect = (*dialect)(nil)

type dialect struct{}

func (dialect) DriverName() string {
	return "postgres"
}

func (c *dialect) indexName(key sqlbuilder.Key) sqlfrag.Fragment {
	name := key.Name()
	if name == "primary" {
		name = "pkey"
	}
	return sqlfrag.Const(sqlbuilder.GetKeyTable(key).TableName() + "_" + name)
}

func (c *dialect) AddIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		return sqlfrag.Pair("\nALTER TABLE ? ADD PRIMARY KEY (?);", sqlbuilder.GetKeyTable(key), sqlbuilder.ColumnCollect(key.Cols()))
	}

	keyDef := key.(sqlbuilder.KeyDef)

	return sqlfrag.Pair("\nCREATE @index_type @index_name ON @table @index_method (@columnAndOptions);", sqlfrag.NamedArgSet{
		"table": sqlbuilder.GetKeyTable(key),
		"index_type": func() sqlfrag.Fragment {
			if key.IsUnique() {
				return sqlfrag.Const("UNIQUE INDEX")
			}
			return sqlfrag.Const("INDEX")
		}(),
		"index_name": c.indexName(key),
		"index_method": func() sqlfrag.Fragment {
			if m := strings.ToUpper(keyDef.Method()); m != "" {
				if m == "SPATIAL" {
					m = "GIST"
				}
				return sqlfrag.Const(fmt.Sprintf("USING %s", m))
			}
			return sqlfrag.Empty()
		}(),
		"columnAndOptions": sqlbuilder.AsKeyColumnsTableDef(key),
	})
}

func (c *dialect) DropIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	if key.IsPrimary() {
		return sqlfrag.Pair("\nALTER TABLE ? DROP CONSTRAINT ?;", sqlbuilder.GetKeyTable(key), c.indexName(key))
	}
	return sqlfrag.Pair("\nDROP INDEX IF EXISTS ?;", c.indexName(key))
}

func (c *dialect) CreateTableIsNotExists(t sqlbuilder.Table) (exprs []sqlfrag.Fragment) {
	exprs = append(exprs, sqlfrag.Pair("\nCREATE TABLE IF NOT EXISTS @table (@def\n);", sqlfrag.NamedArgSet{
		"table": t,
		"def": sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
			return func(yield func(string, []any) bool) {
				idx := 0

				for col := range t.Cols() {
					def := sqlbuilder.GetColumnDef(col)

					// skip deprecated col
					if def.DeprecatedActions != nil {
						continue
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

	return exprs
}

func (c *dialect) DropTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Pair("\nDROP TABLE IF EXISTS @table;", sqlfrag.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) TruncateTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Pair("\nTRUNCATE TABLE @table;", sqlfrag.NamedArgSet{
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
	def := sqlbuilder.GetColumnDef(col)
	prevDef := sqlbuilder.GetColumnDef(prev)

	if def.AutoIncrement {
		return nil
	}

	dbDataType := c.dataType(def.Type, def)
	prevDbDataType := c.dataType(prevDef.Type, prevDef)

	actions := make([]sqlfrag.Fragment, 0)

	if dbDataType != prevDbDataType {
		actions = append(actions, sqlfrag.Pair(
			"ALTER COLUMN ? TYPE ? /* FROM ? */",
			col, sqlfrag.Const(dbDataType), sqlfrag.Const(prevDbDataType),
		))
	}

	if def.Null != prevDef.Null {
		action := "SET"
		if def.Null {
			action = "DROP"
		}

		actions = append(actions, sqlfrag.Pair(
			"ALTER COLUMN ? ? NOT NULL",
			col, sqlfrag.Const(action),
		))
	}

	defaultValue := normalizeDefaultValue(def.Default, dbDataType)
	prevDefaultValue := normalizeDefaultValue(prevDef.Default, prevDbDataType)

	if defaultValue != prevDefaultValue {
		if def.Default != nil {
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
	dbDataType := dealias(c.dbDataType(columnType.Type, columnType))
	return sqlfrag.Pair(dbDataType + autocompleteSize(dbDataType, columnType) + c.dataTypeModify(columnType, dbDataType))
}

func (c *dialect) dataType(typ typex.Type, columnType sqlbuilder.ColumnDef) string {
	dbDataType := dealias(c.dbDataType(typ, columnType))
	return dbDataType + autocompleteSize(dbDataType, columnType)
}

func (c *dialect) dbDataType(typ typex.Type, columnType sqlbuilder.ColumnDef) string {
	if columnType.DataType != "" {
		return columnType.DataType
	}

	if rv, ok := typex.TryNew(typ); ok {
		if dtd, ok := rv.Interface().(sqlbuilder.DataTypeDescriber); ok {
			return dtd.DataType(c.DriverName())
		}
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return c.dataType(typ.Elem(), columnType)
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		if columnType.AutoIncrement {
			return "serial"
		}
		return "integer"
	case reflect.Int64, reflect.Uint64:
		if columnType.AutoIncrement {
			return "bigserial"
		}
		return "bigint"
	case reflect.Float64:
		return "double precision"
	case reflect.Float32:
		return "real"
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "bytea"
		}
	case reflect.String:
		size := columnType.Length
		if size < 65535/3 {
			return "varchar"
		}
		return "text"
	}

	switch typ.Name() {
	case "Hstore":
		return "hstore"
	case "NullInt64":
		return "bigint"
	case "NullFloat64":
		return "double precision"
	case "NullBool":
		return "boolean"
	case "Time", "NullTime":
		return "timestamp with time zone"
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

func normalizeDefaultValue(defaultValue *string, dataType string) string {
	if defaultValue == nil {
		return ""
	}

	dv := *defaultValue

	if dv[0] == '\'' {
		if strings.Contains(dv, "'::") {
			return dv
		}
		return dv + "::" + dataType
	}

	_, err := strconv.ParseFloat(dv, 64)
	if err == nil {
		return "'" + dv + "'::" + dataType
	}

	return dv
}

func autocompleteSize(dataType string, columnType sqlbuilder.ColumnDef) string {
	switch dataType {
	case "character varying", "character":
		size := columnType.Length
		if size == 0 {
			size = 255
		}
		return sizeModifier(size, columnType.Decimal)
	case "decimal", "numeric", "real", "double precision":
		if columnType.Length > 0 {
			return sizeModifier(columnType.Length, columnType.Decimal)
		}
	}
	return ""
}

func dealias(dataType string) string {
	switch dataType {
	case "varchar":
		return "character varying"
	case "timestamp":
		return "timestamp without time zone"
	}
	return dataType
}

func sizeModifier(length uint64, decimal uint64) string {
	if length > 0 {
		size := strconv.FormatUint(length, 10)
		if decimal > 0 {
			return "(" + size + "," + strconv.FormatUint(decimal, 10) + ")"
		}
		return "(" + size + ")"
	}
	return ""
}
