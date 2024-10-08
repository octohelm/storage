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

func IndexColNameAndOptions(colNameAndOptions ...string) IndexOptionFunc {
	return func(k *key) {
		k.colNameAndOptions = colNameAndOptions
	}
}

func Index(name string, columns ColumnCollection, optFns ...IndexOptionFunc) Key {
	k := &key{
		name: strings.ToLower(name),
	}

	if columns != nil {
		for col := range columns.Cols() {
			k.colNameAndOptions = append(k.colNameAndOptions, col.Name())
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
	ColNameAndOptions() []string
}

type key struct {
	table             Table
	name              string
	isUnique          bool
	method            string
	colNameAndOptions []string
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

func (k *key) ColNameAndOptions() []string {
	return k.colNameAndOptions
}

func (k key) Of(table Table) Key {
	return &key{
		table:             table,
		name:              k.name,
		isUnique:          k.isUnique,
		method:            k.method,
		colNameAndOptions: k.colNameAndOptions,
	}
}

func (k *key) Name() string {
	return k.name
}

func (k *key) IsUnique() bool {
	return k.isUnique
}

func (k *key) IsPrimary() bool {
	return k.isUnique && k.name == "primary" || strings.HasSuffix(k.name, "pkey")
}

func (k *key) Cols() iter.Seq[Column] {
	if len(k.colNameAndOptions) == 0 {
		panic(fmt.Errorf("invalid key %s of %s, missing cols", k.name, k.table.TableName()))
	}

	return func(yield func(Column) bool) {
		names := map[string]bool{}
		for _, colNameAndOptions := range k.colNameAndOptions {
			names[strings.Split(colNameAndOptions, "/")[0]] = true
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
