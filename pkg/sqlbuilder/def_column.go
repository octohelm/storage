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

	Name() string
	MatchName(name string) bool

	FieldName() string

	ValueBy(v any) Assignment
	Incr(d int) SqlExpr
	Dec(d int) SqlExpr
	Like(v string) SqlCondition
	LeftLike(v string) SqlCondition
	RightLike(v string) SqlCondition
	NotLike(v string) SqlCondition
	IsNull() SqlCondition
	IsNotNull() SqlCondition
	In(args ...any) SqlCondition
	NotIn(args ...any) SqlCondition

	Between(leftValue any, rightValue any) SqlCondition
	NotBetween(leftValue any, rightValue any) SqlCondition

	Eq(v any) SqlCondition
	Neq(v any) SqlCondition
	Gt(v any) SqlCondition
	Gte(v any) SqlCondition
	Lt(v any) SqlCondition
	Lte(v any) SqlCondition
}

type ColOptionFunc func(c *column)

func ColField(fieldName string) ColOptionFunc {
	return func(c *column) {
		c.fieldName = fieldName
	}
}

func ColDef(def ColumnDef) ColOptionFunc {
	return func(c *column) {
		c.def = def
	}
}

func ColTypeOf(v any, tagValue string) ColOptionFunc {
	return func(c *column) {
		c.def = *ColumnDefFromTypeAndTag(types.FromRType(reflect.TypeOf(v)), tagValue)
	}
}

func Col(name string, fns ...ColOptionFunc) Column {
	c := &column{
		name: strings.ToLower(name),
		def:  ColumnDef{},
	}

	for i := range fns {
		fns[i](c)
	}

	return c
}

var _ TableDefinition = (*column)(nil)

type column struct {
	name      string
	fieldName string
	table     Table
	def       ColumnDef
}

func (c *column) FieldName() string {
	return c.fieldName
}

func (c *column) Def() ColumnDef {
	return c.def
}

func (c *column) With(optionFns ...ColOptionFunc) Column {
	cc := &column{
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

func (c *column) MatchName(name string) bool {
	if name == "" {
		return false
	}

	// first child upper should be fieldName
	if name[0] >= 'A' && name[0] <= 'Z' {
		return c.fieldName == name
	}

	return c.name == name
}

func (c *column) T() Table {
	return c.table
}

func (c *column) Name() string {
	return c.name
}

func (c column) Of(table Table) Column {
	return &column{
		table:     table,
		name:      c.name,
		fieldName: c.fieldName,
		def:       c.def,
	}
}

func (c *column) IsNil() bool {
	return c == nil
}

func (c *column) Ex(ctx context.Context) *Ex {
	toggles := TogglesFromContext(ctx)
	if toggles.Is(ToggleMultiTable) {
		if c.table == nil {
			panic(fmt.Errorf("table of %s is not defined", c.name))
		}
		if toggles.Is(ToggleNeedAutoAlias) {
			return Expr("?.? AS ?", c.table, Expr(c.name), Expr(c.name)).Ex(ctx)
		}
		return Expr("?.?", c.table, Expr(c.name)).Ex(ctx)
	}
	return ExactlyExpr(c.name).Ex(ctx)
}

func (c *column) Expr(query string, args ...any) *Ex {
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

func (c *column) ValueBy(v any) Assignment {
	return ColumnsAndValues(c, v)
}

func (c *column) Incr(d int) SqlExpr {
	return Expr("? + ?", c, d)
}

func (c *column) Dec(d int) SqlExpr {
	return Expr("? - ?", c, d)
}

func (c *column) Like(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, "%"+v+"%"))
}

func (c *column) LeftLike(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, "%"+v))
}

func (c *column) RightLike(v string) SqlCondition {
	return AsCond(Expr("? LIKE ?", c, v+"%"))
}

func (c *column) NotLike(v string) SqlCondition {
	return AsCond(Expr("? NOT LIKE ?", c, "%"+v+"%"))
}

func (c *column) IsNull() SqlCondition {
	return AsCond(Expr("? IS NULL", c))
}

func (c *column) IsNotNull() SqlCondition {
	return AsCond(Expr("? IS NOT NULL", c))
}

type WithConditionFor interface {
	ConditionFor(c Column) SqlCondition
}

func (c *column) In(args ...any) SqlCondition {
	n := len(args)

	switch n {
	case 0:
		return nil
	case 1:
		if withConditionFor, ok := args[0].(WithConditionFor); ok {
			return withConditionFor.ConditionFor(c)
		}
	}

	e := Expr("? IN ")

	e.Grow(n + 1)

	e.AppendArgs(c)

	e.WriteGroup(func(e *Ex) {
		for i := 0; i < n; i++ {
			e.WriteHolder(i)
		}
	})

	e.AppendArgs(args...)

	return AsCond(e)
}

func (c *column) NotIn(args ...any) SqlCondition {
	n := len(args)
	if n == 0 {
		return nil
	}

	e := Expr("")
	e.Grow(n + 1)

	e.WriteQuery("? NOT IN ")
	e.AppendArgs(c)

	e.WriteGroup(func(e *Ex) {
		for i := 0; i < n; i++ {
			e.WriteHolder(i)
		}
	})

	e.AppendArgs(args...)

	return AsCond(e)
}

func (c *column) Between(leftValue any, rightValue any) SqlCondition {
	return AsCond(Expr("? BETWEEN ? AND ?", c, leftValue, rightValue))
}

func (c *column) NotBetween(leftValue any, rightValue any) SqlCondition {
	return AsCond(Expr("? NOT BETWEEN ? AND ?", c, leftValue, rightValue))
}

func (c *column) Eq(v any) SqlCondition {
	return AsCond(Expr("? = ?", c, v))
}

func (c *column) Neq(v any) SqlCondition {
	return AsCond(Expr("? <> ?", c, v))
}

func (c *column) Gt(v any) SqlCondition {
	return AsCond(Expr("? > ?", c, v))
}

func (c *column) Gte(v any) SqlCondition {
	return AsCond(Expr("? >= ?", c, v))
}

func (c *column) Lt(v any) SqlCondition {
	return AsCond(Expr("? < ?", c, v))
}

func (c *column) Lte(v any) SqlCondition {
	return AsCond(Expr("? <= ?", c, v))
}
