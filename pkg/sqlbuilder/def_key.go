package sqlbuilder

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func PrimaryKey(columns ColumnCollection, optFns ...IndexOptionFunc) Key {
	return UniqueIndex("PRIMARY", columns, optFns...)
}

func UniqueIndex(name string, columns ColumnCollection, optFns ...IndexOptionFunc) Key {
	return Index(name, columns, append(optFns, IndexUnique(true))...)
}

type IndexOptionFunc func(k *key)

func IndexUnique(unique bool) IndexOptionFunc {
	return func(k *key) {
		k.isUnique = unique
	}
}

func IndexUsing(method string) IndexOptionFunc {
	return func(k *key) {
		k.method = method
	}
}

func IndexFieldNameAndOptions(colNameAndOptions ...FieldNameAndOption) IndexOptionFunc {
	return func(k *key) {
		k.fieldNameAndOptions = colNameAndOptions
	}
}

func Index(name string, columns ColumnCollection, optFns ...IndexOptionFunc) Key {
	k := &key{
		name: strings.ToLower(name),
	}

	if columns != nil {
		for col := range columns.Cols() {
			k.fieldNameAndOptions = append(k.fieldNameAndOptions, FieldNameAndOption(col.Name()))
		}
	}

	for i := range optFns {
		optFns[i](k)
	}

	return k
}

func GetKeyTable(key Key) Table {
	if withDef, ok := key.(WithTable); ok {
		return withDef.T()
	}
	return nil
}

type Key interface {
	sqlfrag.Fragment

	Of(table Table) Key

	Name() string

	IsPrimary() bool
	IsUnique() bool

	ColumnSeq
}

type KeyDef interface {
	Method() string
	FieldNameAndOptions() []FieldNameAndOption
}

func GetKeyDef(col Key) KeyDef {
	if keyDef, ok := col.(KeyDef); ok {
		return keyDef
	}
	return nil
}

func KeyColumnOnly() func(o *opt) {
	return func(o *opt) {
		o.KeyColumnOnly = true
	}
}

type opt struct {
	KeyColumnOnly bool
}

func AsKeyColumnsTableDef(key Key, optionFns ...func(o *opt)) sqlfrag.Fragment {
	o := &opt{}
	for _, optFn := range optionFns {
		optFn(o)
	}

	keyDef := GetKeyDef(key)
	cc := ColumnCollect(key.Cols())
	fieldNameAndOptions := keyDef.FieldNameAndOptions()

	if len(fieldNameAndOptions) == 0 {
		panic(fmt.Errorf("invalid key %s, missing cols", key))
	}

	return sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
		return func(yield func(string, []any) bool) {
			for i, fo := range fieldNameAndOptions {
				if i > 0 {
					if !yield(",", nil) {
						return
					}
				}

				if c := cc.F(fo.Name()); c != nil {
					if !yield(c.Name(), nil) {
						return
					}

					if !o.KeyColumnOnly {
						if options := fo.Options(); len(options) > 0 {
							if !yield(" "+strings.Join(options, " "), nil) {
								return
							}
							return
						}
					}

					continue
				}

				if c := cc.Col(fo.Name()); c != nil {
					if !yield(c.Name(), nil) {
						return
					}

					if options := fo.Options(); len(options) > 0 {
						if !yield(" "+strings.Join(options, " "), nil) {
							return
						}
						return
					}

					continue
				}

				panic(fmt.Errorf("invalid key %s, unknown %s", key, fo.Name()))
			}
		}
	})
}

type key struct {
	table               Table
	name                string
	isUnique            bool
	method              string
	fieldNameAndOptions []FieldNameAndOption
}

func (k *key) IsNil() bool {
	return k == nil
}

func (k *key) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.Const(k.name).Frag(ctx)
}

func (k *key) T() Table {
	return k.table
}

func (k *key) Method() string {
	return k.method
}

func (k *key) FieldNameAndOptions() []FieldNameAndOption {
	return k.fieldNameAndOptions
}

func (k key) Of(table Table) Key {
	return &key{
		table:               table,
		name:                k.name,
		isUnique:            k.isUnique,
		method:              k.method,
		fieldNameAndOptions: k.fieldNameAndOptions,
	}
}

func (k *key) Name() string {
	return k.name
}

func (k *key) String() string {
	if k.table != nil {
		return fmt.Sprintf("%s.%s", k.table.TableName(), k.name)
	}
	return k.name
}

func (k *key) IsUnique() bool {
	return k.isUnique
}

func (k *key) IsPrimary() bool {
	return k.isUnique && (k.name == "primary" || strings.HasSuffix(k.name, "pkey"))
}

func (k *key) Cols() iter.Seq[Column] {
	if len(k.fieldNameAndOptions) == 0 {
		panic(fmt.Errorf("invalid key %s, missing cols", k))
	}

	return func(yield func(Column) bool) {
		names := map[string]bool{}
		for _, colNameAndOption := range k.fieldNameAndOptions {
			names[colNameAndOption.Name()] = true
		}

		for c := range k.table.Cols() {
			if names[c.Name()] || names[c.FieldName()] {
				if !yield(c) {
					return
				}
			}
		}
	}
}
