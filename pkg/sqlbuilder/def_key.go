package sqlbuilder

import (
	"context"
	"fmt"
	"strings"
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
		columns.RangeCol(func(col Column, idx int) bool {
			k.colNameAndOptions = append(k.colNameAndOptions, col.Name())
			return true
		})
	}

	for i := range optFns {
		optFns[i](k)
	}

	return k
}

type Key interface {
	SqlExpr
	TableDefinition

	Of(table Table) Key

	IsPrimary() bool
	IsUnique() bool
	Name() string
	Columns() ColumnCollection
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

func (k *key) Ex(ctx context.Context) *Ex {
	return ExactlyExpr(k.name).Ex(ctx)
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

func (k *key) Columns() ColumnCollection {
	if len(k.colNameAndOptions) == 0 {
		panic(fmt.Errorf("invalid key %s of %s, missing cols", k.name, k.table.TableName()))
	}

	names := make([]string, len(k.colNameAndOptions))

	for i := range names {
		names[i] = strings.Split(k.colNameAndOptions[i], "/")[0]
	}

	return k.table.Cols(names...)
}
