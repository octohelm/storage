package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/storage/pkg/sqlbuilder"

	"github.com/octohelm/storage/internal/sql/scanner/nullable"
	reflectx "github.com/octohelm/x/reflect"
)

func scanTo(ctx context.Context, rows *sql.Rows, v interface{}) error {
	tpe := reflect.TypeOf(v)

	if tpe.Kind() != reflect.Ptr {
		return fmt.Errorf("scanTo target must be a ptr value, but got %T", v)
	}

	if s, ok := v.(sql.Scanner); ok {
		return rows.Scan(s)
	}

	tpe = reflectx.Deref(tpe)

	switch tpe.Kind() {
	case reflect.Struct:
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		n := len(columns)
		if n < 1 {
			return nil
		}

		dest := make([]interface{}, n)
		holder := placeholder()

		columnIndexes := map[string]int{}

		for i, columnName := range columns {
			columnIndexes[strings.ToLower(columnName)] = i
			dest[i] = holder
		}

		sqlbuilder.ForEachStructFieldValue(ctx, v, func(sf *sqlbuilder.StructFieldValue) {
			if sf.TableName != "" {
				if i, ok := columnIndexes[sf.TableName+"__"+sf.Field.Name]; ok && i > -1 {
					dest[i] = nullable.NewNullIgnoreScanner(sf.Value.Addr().Interface())
				}
			}

			if i, ok := columnIndexes[sf.Field.Name]; ok && i > -1 {
				dest[i] = nullable.NewNullIgnoreScanner(sf.Value.Addr().Interface())
			}
		})

		return rows.Scan(dest...)
	default:
		return rows.Scan(nullable.NewNullIgnoreScanner(v))
	}
}

func placeholder() sql.Scanner {
	p := emptyScanner(0)
	return &p
}

type emptyScanner int

func (e *emptyScanner) Scan(value interface{}) error {
	return nil
}
