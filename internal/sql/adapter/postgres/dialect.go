package postgres

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	typex "github.com/octohelm/x/types"
)

var _ adapter.Dialect = (*dialect)(nil)

type dialect struct {
}

func (dialect) DriverName() string {
	return "postgres"
}

func (c *dialect) indexName(key sqlbuilder.Key) string {
	name := key.Name()
	if name == "primary" {
		name = "pkey"
	}
	return key.T().TableName() + "_" + name
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

	e.WriteQuery(c.indexName(key))

	e.WriteQuery(" ON ")
	e.WriteExpr(key.T())

	keyDef := key.(sqlbuilder.KeyDef)

	if m := strings.ToUpper(keyDef.Method()); m != "" {
		if m == "SPATIAL" {
			m = "GIST"
		}
		e.WriteQuery(" USING ")
		e.WriteQuery(m)
	}

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
		e := sqlbuilder.Expr("ALTER TABLE ")
		e.WriteExpr(key.T())
		e.WriteQuery(" DROP CONSTRAINT ")
		e.WriteQuery(c.indexName(key))
		e.WriteEnd()
		return e
	}
	e := sqlbuilder.Expr("DROP ")

	e.WriteQuery("INDEX IF EXISTS ")
	e.WriteQuery(c.indexName(key))
	e.WriteEnd()

	return e
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

		cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
			def := col.Def()

			if def.DeprecatedActions != nil {
				return true
			}

			if idx > 0 {
				e.WriteQueryByte(',')
			}
			e.WriteQueryByte('\n')
			e.WriteQueryByte('\t')

			e.WriteExpr(col)
			e.WriteQueryByte(' ')
			e.WriteExpr(c.DataType(def))

			return true
		})

		t.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
			if key.IsPrimary() {
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

func (c *dialect) ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlbuilder.SqlExpr {
	def := col.Def()
	prevDef := prev.Def()

	if def.AutoIncrement {
		return nil
	}

	e := sqlbuilder.Expr("ALTER TABLE ")
	e.WriteExpr(col.T())

	dbDataType := c.dataType(def.Type, def)
	prevDbDataType := c.dataType(prevDef.Type, prevDef)

	isFirstSub := true
	isEmpty := true

	prepareAppendSubCmd := func() {
		if !isFirstSub {
			e.WriteQueryByte(',')
		}
		isFirstSub = false
		isEmpty = false
	}

	if dbDataType != prevDbDataType {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		e.WriteQuery(" TYPE ")
		e.WriteQuery(dbDataType)

		e.WriteQuery(" /* FROM ")
		e.WriteQuery(prevDbDataType)
		e.WriteQuery(" */")
	}

	if def.Null != prevDef.Null {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		if !def.Null {
			e.WriteQuery(" SET NOT NULL")
		} else {
			e.WriteQuery(" DROP NOT NULL")
		}
	}

	defaultValue := normalizeDefaultValue(def.Default, dbDataType)
	prevDefaultValue := normalizeDefaultValue(prevDef.Default, prevDbDataType)

	if defaultValue != prevDefaultValue {
		prepareAppendSubCmd()

		e.WriteQuery(" ALTER COLUMN ")
		e.WriteExpr(col)
		if def.Default != nil {
			e.WriteQuery(" SET DEFAULT ")
			e.WriteQuery(defaultValue)

			e.WriteQuery(" /* FROM ")
			e.WriteQuery(prevDefaultValue)
			e.WriteQuery(" */")
		} else {
			e.WriteQuery(" DROP DEFAULT")
		}
	}

	if isEmpty {
		return nil
	}

	e.WriteEnd()

	return e
}

func (c *dialect) DropColumn(col sqlbuilder.Column) sqlbuilder.SqlExpr {
	return sqlbuilder.Expr("ALTER TABLE @table DROP COLUMN @col;", sqlbuilder.NamedArgSet{
		"table": col.T(),
		"col":   col,
	})
}

func (c *dialect) DataType(columnType sqlbuilder.ColumnDef) sqlbuilder.SqlExpr {
	dbDataType := dealias(c.dbDataType(columnType.Type, columnType))
	return sqlbuilder.Expr(dbDataType + autocompleteSize(dbDataType, columnType) + c.dataTypeModify(columnType, dbDataType))
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
