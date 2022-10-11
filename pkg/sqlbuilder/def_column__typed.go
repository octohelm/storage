package sqlbuilder

import (
	"strings"
)

func CastCol[T any](col Column) TypedColumn[T] {
	return &column[T]{
		name:      col.Name(),
		fieldName: col.FieldName(),
		table:     col.T(),
		def:       col.Def(),
	}
}

func TypedColOf[T any](t Table, name string) TypedColumn[T] {
	return CastCol[T](t.F(name))
}

func TypedCol[T any](name string, fns ...ColOptionFunc) TypedColumn[T] {
	c := &column[T]{
		name: strings.ToLower(name),
		def:  ColumnDef{},
	}
	for i := range fns {
		fns[i](c)
	}
	return c
}

type TypedColumn[T any] interface {
	Column

	V(op ColumnValueExpr[T]) SqlExpr
	By(ops ...ColumnValueExpr[T]) Assignment
}

// +gengo:runtimedoc=false
type ColumnValueExpr[T any] func(v Column) SqlExpr

func Value[T any](v T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("?", v)
	}
}

func Incr[T any](v T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? + ?", c, v)
	}
}

func Des[T any](v T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? - ?", c, v)
	}
}

func Eq[T comparable](expect T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? = ?", c, expect)
	}
}

func EqCol[T comparable](expect TypedColumn[T]) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? = ?", c, expect)
	}
}

func NeqCol[T comparable](expect TypedColumn[T]) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? <> ?", c, expect)
	}
}

func In[T any](values ...T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		if len(values) == 0 {
			return nil
		}
		return Expr("? IN (?)", c, values)
	}
}

func NotIn[T any](values ...T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		if len(values) == 0 {
			return nil
		}
		return Expr("? NOT IN (?)", c, values)
	}
}

func IsNull[T any]() ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? IS NULL", c)
	}
}

func IsNotNull[T any]() ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? IS NOT NULL", c)
	}
}

func Neq[T any](expect T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? <> ?", c, expect)
	}
}

func Like[T ~string](s T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? LIKE ?", c, "%"+s+"%")
	}
}

func NotLike[T ~string](s T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? NOT LIKE ?", c, "%"+s+"%")
	}
}

func LeftLike[T ~string](s T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? LIKE ?", c, "%"+s)
	}
}

func RightLike[T ~string](s T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? LIKE ?", c, s+"%")
	}
}

func Between[T comparable](leftValue T, rightValue T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

func NotBetween[T comparable](leftValue T, rightValue T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? NOT BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

func Gt[T comparable](min T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? > ?", c, min)
	}
}

func Gte[T comparable](min T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? >= ?", c, min)
	}
}

func Lt[T comparable](max T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? < ?", c, max)
	}
}

func Lte[T comparable](max T) ColumnValueExpr[T] {
	return func(c Column) SqlExpr {
		return Expr("? <= ?", c, max)
	}
}
