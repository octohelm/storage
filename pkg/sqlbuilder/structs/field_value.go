package structs

import (
	"context"
	"iter"
	"reflect"

	reflectx "github.com/octohelm/x/reflect"
	typesx "github.com/octohelm/x/types"

	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

func AllFieldValueOmitZero(ctx context.Context, v any, excludeFields ...string) iter.Seq[*FieldValue] {
	excludeValues := make(map[string]bool, len(excludeFields))

	for _, fieldName := range excludeFields {
		excludeValues[fieldName] = true
	}

	return func(yield func(*FieldValue) bool) {
		for sfv := range AllFieldValue(ctx, v) {
			if excludeValues[sfv.Field.FieldName] || !reflectx.IsEmptyValue(sfv.Value) {
				if !yield(sfv) {
					return
				}
			}
		}
	}
}

func AllFieldValue(ctx context.Context, v any) iter.Seq[*FieldValue] {
	return func(yield func(*FieldValue) bool) {
		rv, ok := v.(reflect.Value)
		if ok {
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				rv.Set(reflectx.New(rv.Type()))
			}
			v = rv.Interface()
		}

		if m, ok := v.(internal.Model); ok {
			ctx = internal.TableNameContext.Inject(ctx, m.TableName())
		}

		rv = reflectx.Indirect(reflect.ValueOf(v))

		for _, f := range Fields(ctx, typesx.FromRType(reflect.TypeOf(v))) {
			tagDB := f.Tags["db"]

			if tagDB.HasFlag("deprecated") {
				continue
			}

			if tableAlias, ok := f.Tags["alias"]; ok {
				ctx = internal.TableAliasContext.Inject(ctx, tableAlias.Name())
			} else {
				if len(f.ModelLoc) > 0 {
					fpv := f.FieldModelValue(rv)
					if fpv.IsValid() {
						if m, ok := fpv.Interface().(internal.Model); ok {
							ctx = internal.TableNameContext.Inject(ctx, m.TableName())
						}
					}
				}
			}

			sf := &FieldValue{}

			sf.Field = *f
			sf.Value = f.FieldValue(rv)

			if tableName, ok := internal.TableNameContext.MayFrom(ctx); ok && tableName != "" {
				sf.TableName = tableName
			}

			if tableAlias, ok := internal.TableAliasContext.MayFrom(ctx); ok && tableAlias != "" {
				sf.TableName = tableAlias
			}

			if !yield(sf) {
				return
			}
		}
	}
}

type FieldValue struct {
	Field     Field
	TableName string
	Value     reflect.Value
}
