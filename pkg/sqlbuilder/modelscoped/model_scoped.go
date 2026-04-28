package modelscoped

import (
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

// M 为 modelscoped 包装类型提供类型化模型分配能力。
type M[Model internal.Model] struct{}

// Model 返回一个新的零值模型实例。
func (M[Model]) Model() *Model {
	return new(Model)
}

// ModelNewer 复用通用的类型化模型工厂约束。
type ModelNewer[Model internal.Model] internal.ModelNewer[Model]

// FromModel 根据模型元数据返回类型化表包装。
func FromModel[Model internal.Model]() Table[Model] {
	return CastTable[Model](sqlbuilder.TableFromModel(new(Model)))
}

// CastTable 把无类型的 sqlbuilder.Table 包装为类型化表。
func CastTable[Model internal.Model](t sqlbuilder.Table) Table[Model] {
	return &table[Model]{
		Table: t,
	}
}

// Table 是 sqlbuilder.Table 的类型化视图。
type Table[Model internal.Model] interface {
	sqlbuilder.Table

	ModelNewer[Model]

	MK(key string) Key[Model]

	ColumnSeq[Model]

	KeySeq[Model]
}

// ColumnSeq 是类型化列迭代器。
type ColumnSeq[Model internal.Model] interface {
	sqlbuilder.ColumnSeq

	MCols() iter.Seq[Column[Model]]
}

// KeySeq 是类型化索引迭代器。
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

// CastKey 把无类型索引包装为类型化索引。
func CastKey[Model internal.Model](k sqlbuilder.Key) Key[Model] {
	return &key[Model]{
		Key: k,
	}
}

// Key 是 sqlbuilder.Key 的类型化视图。
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

// CastColumn 把无类型列包装为类型化列。
func CastColumn[Model internal.Model](col sqlbuilder.Column) Column[Model] {
	return &column[Model]{
		Column: col,
	}
}

// AllColumns 按原始顺序产出列。
func AllColumns[Model internal.Model](columns ...Column[Model]) iter.Seq[Column[Model]] {
	return func(yield func(Column[Model]) bool) {
		for _, col := range columns {
			if !yield(col) {
				return
			}
		}
	}
}

// Column 是 sqlbuilder.Column 的类型化视图。
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

// CastTypedColumn 把无类型列包装为类型化的 TypedColumn。
func CastTypedColumn[Model internal.Model, T any](col sqlbuilder.Column) TypedColumn[Model, T] {
	return &typedColumn[Model, T]{
		TypedColumn: sqlbuilder.CastColumn[T](col),
	}
}

// TypedCol 按列名创建类型化列。
func TypedCol[Model internal.Model, T any](name string, opts ...sqlbuilder.ColOptionFunc) TypedColumn[Model, T] {
	return &typedColumn[Model, T]{
		TypedColumn: sqlbuilder.TypedCol[T](name, opts...),
	}
}

// TypedColumn 表示带类型化取值辅助能力的列。
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
