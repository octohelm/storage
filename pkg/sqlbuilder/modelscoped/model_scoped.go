package modelscoped

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"iter"
)

type M[Model internal.Model] struct {
}

func (M[Model]) Model() *Model {
	return new(Model)
}

type ModelNewer[Model internal.Model] internal.ModelNewer[Model]

func FromModel[Model internal.Model]() Table[Model] {
	return CastTable[Model](sqlbuilder.TableFromModel(new(Model)))
}

func CastTable[Model internal.Model](t sqlbuilder.Table) Table[Model] {
	return &table[Model]{
		Table: t,
	}
}

type Table[Model internal.Model] interface {
	sqlbuilder.Table

	ModelNewer[Model]

	MK(key string) Key[Model]

	ColumnSeq[Model]

	KeySeq[Model]
}

type ColumnSeq[Model internal.Model] interface {
	sqlbuilder.ColumnSeq

	MCols() iter.Seq[Column[Model]]
}

type KeySeq[Model internal.Model] interface {
	MKeys() iter.Seq[Key[Model]]
}

type table[Model internal.Model] struct {
	M[Model]

	sqlbuilder.Table
}

func (t *table[Model]) Unwrap() sqlbuilder.Table {
	return t.Table
}

func (t *table[Model]) MK(name string) Key[Model] {
	return CastKey[Model](t.K(name))
}

func (t *table[Model]) MKeys() iter.Seq[Key[Model]] {
	return func(yield func(Key[Model]) bool) {
		for col := range t.Keys() {
			if !yield(CastKey[Model](col)) {
				return
			}
		}
	}
}

func (t *table[Model]) MCols() iter.Seq[Column[Model]] {
	return func(yield func(Column[Model]) bool) {
		for col := range t.Cols() {
			if !yield(CastColumn[Model](col)) {
				return
			}
		}
	}
}

func CastKey[Model internal.Model](k sqlbuilder.Key) Key[Model] {
	return &key[Model]{
		Key: k,
	}
}

type Key[Model internal.Model] interface {
	ModelNewer[Model]

	sqlbuilder.Key

	ColumnSeq[Model]
}

type key[Model internal.Model] struct {
	M[Model]

	sqlbuilder.Key
}

func (k *key[Model]) MCols() iter.Seq[Column[Model]] {
	return func(yield func(Column[Model]) bool) {
		for col := range k.Cols() {
			if !yield(CastColumn[Model](col)) {
				return
			}
		}
	}
}

func CastColumn[Model internal.Model](col sqlbuilder.Column) Column[Model] {
	return &column[Model]{
		Column: col,
	}
}

func AllColumns[Model internal.Model](columns ...Column[Model]) iter.Seq[Column[Model]] {
	return func(yield func(Column[Model]) bool) {
		for _, col := range columns {
			if !yield(col) {
				return
			}
		}
	}
}

type Column[Model internal.Model] interface {
	ModelNewer[Model]
	sqlbuilder.Column
	ComputedBy(frag sqlfrag.Fragment) Column[Model]
}

type column[Model internal.Model] struct {
	M[Model]

	sqlbuilder.Column
}

func (c *column[M]) Unwrap() sqlbuilder.Column {
	return c.Column
}

func (c *column[Model]) ComputedBy(aggregate sqlfrag.Fragment) Column[Model] {
	return CastColumn[Model](
		sqlbuilder.CastColumn[any](c, sqlbuilder.ColComputedBy(aggregate)),
	)
}

func CastTypedColumn[Model internal.Model, T any](col sqlbuilder.Column) TypedColumn[Model, T] {
	return &typedColumn[Model, T]{
		TypedColumn: sqlbuilder.CastColumn[T](col),
	}
}

func TypedCol[Model internal.Model, T any](name string, opts ...sqlbuilder.ColOptionFunc) TypedColumn[Model, T] {
	return &typedColumn[Model, T]{
		TypedColumn: sqlbuilder.TypedCol[T](name, opts...),
	}
}

type TypedColumn[Model internal.Model, T any] interface {
	ModelNewer[Model]
	sqlbuilder.TypedColumn[T]

	ComputedBy(frag sqlfrag.Fragment) Column[Model]
	TypedComputedBy(frag sqlfrag.Fragment) TypedColumn[Model, T]
}

type typedColumn[Model internal.Model, T any] struct {
	M[Model]

	sqlbuilder.TypedColumn[T]
}

func (c *typedColumn[M, T]) Unwrap() sqlbuilder.Column {
	return c.TypedColumn
}

func (c *typedColumn[Model, T]) ComputedBy(aggregate sqlfrag.Fragment) Column[Model] {
	return CastColumn[Model](
		sqlbuilder.CastColumn[any](c, sqlbuilder.ColComputedBy(aggregate)),
	)
}

func (c *typedColumn[Model, T]) TypedComputedBy(aggregate sqlfrag.Fragment) TypedColumn[Model, T] {
	return CastTypedColumn[Model, T](
		sqlbuilder.CastColumn[T](c, sqlbuilder.ColComputedBy(aggregate)),
	)
}
