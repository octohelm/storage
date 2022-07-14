package sqlite

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	typex "github.com/octohelm/x/types"
)

var _ adapter.Dialect = (*dialect)(nil)

type dialect struct {
}

func (dialect) DriverName() string {
	return "sqlite"
}

func (c *dialect) AddIndex(key sqlbuilder.Key) sqlbuilder.SqlExpr {
	if key.IsPrimary() {
		e := sqlbuilder.Expr("ALTER TABLE ")
		e.WriteExpr(key.T())
		e.WriteQuery(" ADD PRIMARY KEY ")
		e.WriteGroup(func(e *sqlbuilder.Ex) {
			e.WriteExpr(key.Columns())
		})
		e.WriteEnd()
		return e
	}

	e := sqlbuilder.Expr("CREATE ")
	if key.IsUnique() {
		e.WriteQuery("UNIQUE ")
	}
	e.WriteQuery("INDEX ")

	e.WriteExpr(c.indexName(key))

	e.WriteQuery(" ON ")
	e.WriteExpr(key.T())

	keyDef := key.(sqlbuilder.KeyDef)

	e.WriteQueryByte(' ')
	e.WriteGroup(func(e *sqlbuilder.Ex) {
		for i, colNameAndOpt := range keyDef.ColNameAndOptions() {
			parts := strings.Split(colNameAndOpt, "/")
			if i != 0 {
				_ = e.WriteByte(',')
			}
			e.WriteExpr(key.T().F(parts[0]))
			if len(parts) > 1 {
				e.WriteQuery(" ")
				e.WriteQuery(parts[1])
			}
		}
	})

	e.WriteEnd()
	return e
}

func (c *dialect) DropIndex(key sqlbuilder.Key) sqlbuilder.SqlExpr {
	if key.IsPrimary() {
		// pk could not changed
		return nil
	}

	return sqlbuilder.Expr("DROP INDEX IF EXISTS @index;", sqlbuilder.NamedArgSet{
		"index": c.indexName(key),
	})
}

func (c *dialect) CreateTableIsNotExists(t sqlbuilder.Table) (exprs []sqlbuilder.SqlExpr) {
	expr := sqlbuilder.Expr("CREATE TABLE IF NOT EXISTS @table ", sqlbuilder.NamedArgSet{
		"table": t,
	})

	expr.WriteGroup(func(e *sqlbuilder.Ex) {
		cols := t.Cols()

		if cols.IsNil() {
			return
		}

		var autoIncrement sqlbuilder.Column

		cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
			def := col.Def()

			if def.DeprecatedActions != nil {
				return true
			}

			if def.AutoIncrement {
				autoIncrement = col
			}

			if idx > 0 {
				e.WriteQueryByte(',')
			}
			e.WriteQueryByte('\n')
			e.WriteQueryByte('\t')

			e.WriteExpr(col)
			e.WriteQueryByte(' ')
			e.WriteExpr(c.DataType(col.Def()))

			return true
		})

		t.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
			if key.IsPrimary() {
				var skip = false

				if autoIncrement != nil {
					key.Columns().RangeCol(func(col sqlbuilder.Column, idx int) bool {
						if autoIncrement.Name() == col.Name() {
							skip = true
							// auto increment pk will create when table define
							return false
						}
						return true
					})
				}

				if skip {
					return true
				}

				e.WriteQueryByte(',')
				e.WriteQueryByte('\n')
				e.WriteQueryByte('\t')
				e.WriteQuery("PRIMARY KEY ")
				e.WriteGroup(func(e *sqlbuilder.Ex) {
					e.WriteExpr(key.Columns())
				})
			}

			return true
		})

		expr.WriteQueryByte('\n')
	})

	expr.WriteEnd()
	exprs = append(exprs, expr)

	t.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
		if !key.IsPrimary() {
			exprs = append(exprs, c.AddIndex(key))
		}
		return true
	})

	return
}

func (c *dialect) DropTable(t sqlbuilder.Table) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("DROP TABLE IF EXISTS @table;", sqlbuilder.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) TruncateTable(t sqlbuilder.Table) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("TRUNCATE TABLE @table;", sqlbuilder.NamedArgSet{
		"table": t,
	})
}

func (c *dialect) AddColumn(col sqlbuilder.Column) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("ALTER TABLE @table ADD COLUMN @col @dataType;", sqlbuilder.NamedArgSet{
		"table":    col.T(),
		"col":      col,
		"dataType": c.DataType(col.Def()),
	})
}

func (c *dialect) RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("ALTER TABLE @table RENAME COLUMN @oldCol TO @newCol;", sqlbuilder.NamedArgSet{
		"table":  col.T(),
		"oldCol": col,
		"newCol": target,
	})
}

func (c *dialect) ModifyColumn(col sqlbuilder.Column, prevCol sqlbuilder.Column) sqlbuilder.SqlExpr {
	def := col.Def()

	// incr id never modified
	if def.AutoIncrement {
		return nil
	}

	prevTmpCol := sqlbuilder.Col("__"+prevCol.Name(), sqlbuilder.ColDef(prevCol.Def())).Of(prevCol.T())

	e := sqlbuilder.Expr("")

	e.WriteExpr(sqlbuilder.Expr("ALTER TABLE @table RENAME COLUMN @prevCol TO @tmpCol;", sqlbuilder.NamedArgSet{
		"table":   prevCol.T(),
		"prevCol": prevCol,
		"tmpCol":  prevTmpCol,
	}))

	e.WriteExpr(c.AddColumn(col))

	e.WriteExpr(sqlbuilder.Expr("UPDATE @table SET @col = @tmpCol;", sqlbuilder.NamedArgSet{
		"table":  col.T(),
		"col":    col,
		"tmpCol": prevTmpCol,
	}))

	e.WriteExpr(c.DropColumn(prevTmpCol))

	return e
}

func (c *dialect) DropColumn(col sqlbuilder.Column) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("ALTER TABLE @table DROP COLUMN @col;", sqlbuilder.NamedArgSet{
		"table": col.T(),
		"col":   col,
	})
}

func (c *dialect) DataType(columnType sqlbuilder.ColumnDef) sqlbuilder.SqlExpr {
	dbDataType := c.dbDataType(columnType.Type, columnType)
	return sqlbuilder.Expr(dbDataType + c.dataTypeModify(columnType, dbDataType))
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

func (c dialect) indexName(key sqlbuilder.Key) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr(fmt.Sprintf("%s_%s", key.T().TableName(), key.Name()))
}

func normalizeDefaultValue(defaultValue *string, dataType string) string {
	if defaultValue == nil {
		return ""
	}
	return *defaultValue
}
