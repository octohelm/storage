package sqlbuilder

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/x/types"
)

type Column interface {
	SqlExpr
	TableDefinition
	Def() ColumnDef
	Expr(query string, args ...any) *Ex
	Of(table Table) Column
	With(optionFns ...ColOptionFunc) Column
	MatchName(name string) bool
	Name() string
	FieldName() string
}

type ColumnSetter interface {
	SetFieldName(name string)
	SetColumnDef(def ColumnDef)
}

type ColOptionFunc func(c ColumnSetter)

func ColField(fieldName string) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetFieldName(fieldName)
	}
}

func ColDef(def ColumnDef) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetColumnDef(def)
	}
}

func ColTypeOf(v any, tagValue string) ColOptionFunc {
	return func(c ColumnSetter) {
		c.SetColumnDef(*ColumnDefFromTypeAndTag(types.FromRType(reflect.TypeOf(v)), tagValue))
	}
}

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

var _ TableDefinition = (*column[any])(nil)

type column[T any] struct {
	name      string
	fieldName string
	table     Table
	def       ColumnDef
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

func (c *column[T]) With(optionFns ...ColOptionFunc) Column {
	cc := &column[T]{
		name:      c.name,
		fieldName: c.fieldName,
		table:     c.table,
		def:       c.def,
	}

	for i := range optionFns {
		optionFns[i](c)
	}

	return cc
}

func (c *column[T]) MatchName(name string) bool {
	if name == "" {
		return false
	}

	// first child upper should be fieldName
	if name[0] >= 'A' && name[0] <= 'Z' {
		return c.fieldName == name
	}

	return c.name == name
}

func (c *column[T]) T() Table {
	return c.table
}

func (c *column[T]) Name() string {
	return c.name
}

func (c column[T]) Of(table Table) Column {
	return &column[T]{
		table:     table,
		name:      c.name,
		fieldName: c.fieldName,
		def:       c.def,
	}
}

func (c *column[T]) IsNil() bool {
	return c == nil
}

func (c *column[T]) Ex(ctx context.Context) *Ex {
	toggles := TogglesFromContext(ctx)
	if toggles.Is(ToggleMultiTable) {
		if c.table == nil {
			panic(fmt.Errorf("table of %s is not defined", c.name))
		}
		if toggles.Is(ToggleNeedAutoAlias) {
			return Expr("?.? AS ?", c.table, Expr(c.name), Expr(fmt.Sprintf("%s__%s", c.table.TableName(), c.name))).Ex(ctx)
		}
		return Expr("?.?", c.table, Expr(c.name)).Ex(ctx)
	}
	return ExactlyExpr(c.name).Ex(ctx)
}

func (c *column[T]) Expr(query string, args ...any) *Ex {
	n := len(args)
	e := Expr("")
	e.Grow(n)

	qc := 0

	for _, key := range []byte(query) {
		switch key {
		case '#':
			e.WriteExpr(c)
		case '?':
			e.WriteQueryByte(key)
			if n > qc {
				e.AppendArgs(args[qc])
				qc++
			}
		default:
			e.WriteQueryByte(key)
		}
	}

	return e
}

func (c *column[T]) By(ops ...ColumnValueExpr[T]) Assignment {
	if len(ops) == 0 {
		return nil
	}
	values := make([]any, len(ops))
	for i := range ops {
		values[i] = ops[i](c)
	}
	return ColumnsAndValues(c, values...)
}

func (c *column[T]) V(operator ColumnValueExpr[T]) SqlExpr {
	return operator(c)
}
