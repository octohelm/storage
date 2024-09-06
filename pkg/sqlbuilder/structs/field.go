package structs

import (
	"context"
	"database/sql/driver"
	"fmt"
	"go/ast"
	"iter"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal/columndef"
	reflectx "github.com/octohelm/x/reflect"
	typesx "github.com/octohelm/x/types"
)

func Fields(ctx context.Context, typ typesx.Type) []*Field {
	return defaultStructFieldsFactory.TableFieldsFor(ctx, typ)
}

var defaultStructFieldsFactory = &fieldsFactory{}

type fieldsFactory struct {
	cache sync.Map
}

func (tf *fieldsFactory) TableFieldsFor(ctx context.Context, typ typesx.Type) []*Field {
	typ = typesx.Deref(typ)

	underlying := typ.Unwrap()

	if v, ok := tf.cache.Load(underlying); ok {
		return v.([]*Field)
	}

	tfs := make([]*Field, 0)

	for f := range AllStructField(ctx, typ) {
		tagDB := f.Tags["db"]
		if tagDB != "" && tagDB != "-" {
			tfs = append(tfs, f)
		}
	}

	tf.cache.Store(underlying, tfs)

	return tfs
}

func AllStructField(ctx context.Context, tpe typesx.Type) iter.Seq[*Field] {
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}

	w := &fieldWalker{}

	return w.allStructField(ctx, tpe)
}

type fieldWalker struct {
	loc       []int
	modelLoc  []int
	modelType typesx.Type
}

var typeModel = reflect.TypeOf((*internal.Model)(nil)).Elem()
var driverValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

func (w *fieldWalker) allStructField(ctx context.Context, tpe typesx.Type) iter.Seq[*Field] {
	modelLoc := w.modelLoc[:]
	modelType := w.modelType

	if ok := tpe.Implements(typesx.FromRType(typeModel)); ok {
		if modelType != nil && modelType.NumField() == 1 && modelType.Field(0).Anonymous() {
			// extendable

		} else {
			modelType = tpe
			modelLoc = w.loc[:]
		}
	}

	return func(yield func(*Field) bool) {
		for i := 0; i < tpe.NumField(); i++ {
			f := tpe.Field(i)

			if !ast.IsExported(f.Name()) {
				continue
			}

			loc := append(w.loc, i)

			tags := reflectx.ParseStructTags(string(f.Tag()))
			displayName := f.Name()

			tagDB, hasDB := tags["db"]
			if hasDB {
				if name := tagDB.Name(); name == "-" {
					// skip name:"-"
					continue
				} else {
					if name != "" {
						displayName = name
					}
				}
			}

			if (f.Anonymous() || f.Type().Name() == f.Name()) && (!hasDB) {
				fieldType := f.Type()

				if !fieldType.Implements(typesx.FromRType(driverValuer)) {
					for fieldType.Kind() == reflect.Ptr {
						fieldType = fieldType.Elem()
					}

					if fieldType.Kind() == reflect.Struct {
						embed := &fieldWalker{
							loc:       loc,
							modelType: modelType,
							modelLoc:  modelLoc,
						}

						for c := range embed.allStructField(ctx, fieldType) {
							if !yield(c) {
								return
							}
						}

						continue
					}
				}
			}

			p := &Field{}
			p.FieldName = f.Name()
			p.Type = f.Type()
			p.Field = f
			p.Tags = tags
			p.Name = strings.ToLower(displayName)

			p.Loc = make([]int, len(loc))
			copy(p.Loc, loc)

			p.ModelLoc = make([]int, len(modelLoc))
			copy(p.ModelLoc, modelLoc)

			p.ColumnType = *columndef.FromTypeAndTag(p.Type, string(tagDB))

			if !yield(p) {
				return
			}
		}
	}
}

type Field struct {
	Name       string
	FieldName  string
	Type       typesx.Type
	Field      typesx.StructField
	Tags       map[string]reflectx.StructTag
	Loc        []int
	ModelLoc   []int
	ColumnType columndef.ColumnDef
}

func (p *Field) FieldValue(structReflectValue reflect.Value) reflect.Value {
	return fieldValue(structReflectValue, p.Loc)
}

func (p *Field) FieldModelValue(structReflectValue reflect.Value) reflect.Value {
	return fieldValue(structReflectValue, p.ModelLoc)
}

func fieldValue(structReflectValue reflect.Value, locs []int) reflect.Value {
	n := len(locs)

	if n == 0 {
		return structReflectValue
	}

	if n < 0 {
		return reflect.Value{}
	}

	structReflectValue = reflectx.Indirect(structReflectValue)

	fv := structReflectValue

	for i := 0; i < n; i++ {
		loc := locs[i]
		fv = fv.Field(loc)

		// last loc should keep ptr value
		if i < n-1 {
			for fv.Kind() == reflect.Ptr {
				// notice the ptr struct ensure only for Ptr Anonymous Field
				if fv.IsNil() {
					fv.Set(reflectx.New(fv.Type()))
				}
				fv = fv.Elem()
			}
		}
	}

	return fv
}
