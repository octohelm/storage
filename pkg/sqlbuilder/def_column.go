package sqlbuilder

import (
	"context"
	"database/sql"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/octohelm/x/types"

	"github.com/octohelm/storage/pkg/sqlbuilder/internal/columndef"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

// Column 表示 SQL 构建中可引用、可投影的列。
type Column interface {
	sqlfrag.Fragment

	Fragment(query string, args ...any) sqlfrag.Fragment
	Of(table Table) Column
	Name() string
	FieldName() string
}

// GetColumnTable 返回列绑定的表。
func GetColumnTable(col Column) Table {
	if w, ok := col.(ColumnWrapper); ok {
		col = w.Unwrap()
	}
	if withDef, ok := col.(WithTable); ok {
		return withDef.T()
	}
	return nil
}

// ColumnWithDef 暴露列定义元信息。
type ColumnWithDef interface {
	Def() ColumnDef
}

// GetColumnDef 返回列绑定的列定义。
func GetColumnDef(col Column) ColumnDef {
	if w, ok := col.(ColumnWrapper); ok {
		col = w.Unwrap()
	}
	if withDef, ok := col.(ColumnWithDef); ok {
		return withDef.Def()
	}
	return ColumnDef{}
}

// ColumnWithComputed 暴露列的计算表达式。
type ColumnWithComputed interface {
	Computed() sqlfrag.Fragment
}

// GetColumnComputed 返回列绑定的计算表达式。
func GetColumnComputed(col Column) sqlfrag.Fragment {
	if w, ok := col.(ColumnWrapper); ok {
		col = w.Unwrap()
	}
	if withDef, ok := col.(ColumnWithComputed); ok {
		return withDef.Computed()
	}
	return nil
}

// MatchColumn 根据字段名或列名判断是否匹配。
func MatchColumn(col Column, name string) bool {
	if name == "" {
		return false
	}

	// first child upper should be fieldName
	if name[0] >= 'A' && name[0] <= 'Z' {
		return col.FieldName() == name
	}

	return col.Name() == name
}

// ColumnSetter 定义列可被选项函数修改的部分。
type ColumnSetter interface {
	SetFieldName(name string)
	SetColumnDef(def ColumnDef)
	SetComputed(computed sqlfrag.Fragment)
}

// ColOptionFunc 定义列选项函数。
type ColOptionFunc func(c ColumnSetter)

// ColField 设置列对应的结构字段名。
func ColField(fieldName string) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetFieldName(fieldName)
	}
}

// ColComputedBy 设置列的计算表达式。
func ColComputedBy(aggregate sqlfrag.Fragment) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetComputed(aggregate)
	}
}

// ColDef 设置列定义。
func ColDef(def ColumnDef) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetColumnDef(def)
	}
}

// ColTypeOf 根据值类型和标签推导列定义。
func ColTypeOf(v any, tagValue string) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetColumnDef(
			*columndef.FromTypeAndTag(types.FromRType(reflect.TypeOf(v)), tagValue, ""),
		)
	}
}

// Col 按名称创建一个基础列。
func Col(name string, fns ...ColOptionFunc) Column {
	c := &column[any]{
		name: strings.ToLower(name),
		def:  ColumnDef{},
	}

	for i := range fns {
		fns[i](c)
	}

	return c
}

type column[T any] struct {
	name      string
	fieldName string
	def       ColumnDef
	table     Table
	computed  sqlfrag.Fragment
}

func (c *column[T]) SetComputed(computed sqlfrag.Fragment) {
	c.computed = computed
}

func (c *column[T]) SetFieldName(name string) {
	c.fieldName = name
}

func (c *column[T]) SetColumnDef(def ColumnDef) {
	c.def = def
}

func (c *column[T]) FieldName() string {
	return c.fieldName
}

func (c *column[T]) Def() ColumnDef {
	return c.def
}

func (c *column[T]) T() Table {
	return c.table
}

func (c *column[T]) Computed() sqlfrag.Fragment {
	return c.computed
}

func (c *column[T]) Name() string {
	return c.name
}

func (c *column[T]) String() string {
	if c.table != nil {
		return fmt.Sprintf("%s.%s", c.table, c.name)
	}
	return c.name
}

func (c column[T]) Of(table Table) Column {
	return &column[T]{
		table: table,

		name:      c.name,
		fieldName: c.fieldName,
		def:       c.def,
		computed:  c.computed,
	}
}

func (c *column[T]) IsNil() bool {
	return c == nil
}

func (c *column[T]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	toggles := TogglesFromContext(ctx)

	if c.computed != nil && toggles.Is(ToggleInProject) {
		return sqlfrag.Pair("? AS ?", c.computed, sqlfrag.Const(c.name)).Frag(ctx)
	}

	if toggles.Is(ToggleMultiTable) {
		if c.table == nil {
			panic(fmt.Errorf("table of %s is not defined", c.name))
		}

		if toggles.Is(ToggleNeedAutoAlias) {
			return sqlfrag.Pair(
				"?.? AS ?",
				c.table,
				sqlfrag.Const(c.name),
				sqlfrag.Pair(sqlfrag.SafeProjected(c.table.TableName(), c.name)),
			).Frag(ctx)
		}

		return sqlfrag.Pair("?.?", c.table, sqlfrag.Const(c.name)).Frag(ctx)
	}

	return sqlfrag.Const(c.name).Frag(ctx)
}

func (c *column[T]) Expr(query string, args ...any) sqlfrag.Fragment {
	return c.Fragment(query, args...)
}

func (c *column[T]) Fragment(query string, args ...any) sqlfrag.Fragment {
	q := strings.ReplaceAll(query, "#", "@_column")

	return sqlfrag.Pair(q, append([]any{sql.Named("_column", c)}, args)...)
}

func (c *column[T]) By(ops ...ColumnValuer[T]) Assignment {
	if len(ops) == 0 {
		return nil
	}
	values := make([]any, 0, len(ops))
	for _, op := range ops {
		if op == nil {
			continue
		}
		values = append(values, op(c))
	}
	return ColumnsAndValues(c, values...)
}

func (c *column[T]) V(operator ColumnValuer[T]) sqlfrag.Fragment {
	if operator == nil {
		return nil
	}
	return operator(c)
}
