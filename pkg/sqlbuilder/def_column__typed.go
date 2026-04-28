package sqlbuilder

import (
	"iter"
	"strings"

	"github.com/octohelm/storage/internal/xiter"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

// CastColumn 把列包装为指定值类型的 TypedColumn。
func CastColumn[T any](col Column, fns ...ColOptionFunc) TypedColumn[T] {
	c := &column[T]{
		name:      col.Name(),
		fieldName: col.FieldName(),

		table:    GetColumnTable(col),
		def:      GetColumnDef(col),
		computed: GetColumnComputed(col),
	}

	for _, fn := range fns {
		fn(c)
	}

	return c
}

// ColumnWrapper 暴露底层未包装的列。
type ColumnWrapper interface {
	Unwrap() Column
}

// TypedColOf 从指定表中按名称取出并包装为 TypedColumn。
func TypedColOf[T any](t Table, name string) TypedColumn[T] {
	return CastColumn[T](t.F(name))
}

// TypedCol 按名称创建一个新的 TypedColumn。
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

// TypedColumn 表示带类型化取值与赋值辅助方法的列。
type TypedColumn[T any] interface {
	Column

	V(op ColumnValuer[T]) sqlfrag.Fragment
	By(ops ...ColumnValuer[T]) Assignment
}

// +gengo:runtimedoc=false
// ColumnValuer 根据列构造对应的 SQL 片段。
type ColumnValuer[T any] func(v Column) sqlfrag.Fragment

// AsValue 使用另一列作为当前列的赋值来源。
func AsValue[T any](v TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("?", v)
	}
}

// Value 使用字面值作为当前列的赋值来源。
func Value[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("?", v)
	}
}

// Incr 生成当前列加上指定值的表达式。
func Incr[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? + ?", c, v)
	}
}

// Des 生成当前列减去指定值的表达式。
func Des[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? - ?", c, v)
	}
}

// Eq 生成当前列等于指定值的条件。
func Eq[T comparable](expect T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? = ?", c, expect)
	}
}

// EqCol 生成当前列等于另一列的条件。
func EqCol[T comparable](expect TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? = ?", c, expect)
	}
}

// NeqCol 生成当前列不等于另一列的条件。
func NeqCol[T comparable](expect TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <> ?", c, expect)
	}
}

// In 生成当前列属于给定值集合的条件。
func In[T any](values ...T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		if len(values) == 0 {
			return nil
		}
		return sqlfrag.Pair("? IN (?)", c, values)
	}
}

// InSeq 生成当前列属于给定序列的条件。
func InSeq[T any](values iter.Seq[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IN (?)", c, xiter.Map(values, func(x T) any {
			return x
		}))
	}
}

// NotIn 生成当前列不属于给定值集合的条件。
func NotIn[T any](values ...T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		if len(values) == 0 {
			return nil
		}

		return sqlfrag.Pair("? NOT IN (?)", c, values)
	}
}

// NotInSeq 生成当前列不属于给定序列的条件。
func NotInSeq[T any](values iter.Seq[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT IN (?)", c, xiter.Map(values, func(x T) any {
			return x
		}))
	}
}

// IsNull 生成当前列为 NULL 的条件。
func IsNull[T any]() ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IS NULL", c)
	}
}

// IsNotNull 生成当前列不为 NULL 的条件。
func IsNotNull[T any]() ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IS NOT NULL", c)
	}
}

// Neq 生成当前列不等于指定值的条件。
func Neq[T any](expect T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <> ?", c, expect)
	}
}

// Like 生成包含匹配的 LIKE 条件。
func Like[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, "%"+s+"%")
	}
}

// NotLike 生成排除匹配的 NOT LIKE 条件。
func NotLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT LIKE ?", c, "%"+s+"%")
	}
}

// LeftLike 生成左模糊匹配条件。
func LeftLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, "%"+s)
	}
}

// RightLike 生成右模糊匹配条件。
func RightLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, s+"%")
	}
}

// Between 生成闭区间范围条件。
func Between[T comparable](leftValue T, rightValue T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

// NotBetween 生成不在闭区间范围内的条件。
func NotBetween[T comparable](leftValue T, rightValue T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

// Gt 生成大于指定值的条件。
func Gt[T comparable](min T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? > ?", c, min)
	}
}

// Gte 生成大于等于指定值的条件。
func Gte[T comparable](min T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? >= ?", c, min)
	}
}

// Lt 生成小于指定值的条件。
func Lt[T comparable](max T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? < ?", c, max)
	}
}

// Lte 生成小于等于指定值的条件。
func Lte[T comparable](max T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <= ?", c, max)
	}
}
