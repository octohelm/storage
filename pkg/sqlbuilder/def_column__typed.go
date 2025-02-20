package sqlbuilder

import (
	"iter"
	"strings"

	"github.com/octohelm/storage/internal/xiter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

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

type ColumnWrapper interface {
	Unwrap() Column
}

func TypedColOf[T any](t Table, name string) TypedColumn[T] {
	return CastColumn[T](t.F(name))
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

	V(op ColumnValuer[T]) sqlfrag.Fragment
	By(ops ...ColumnValuer[T]) Assignment
}

// +gengo:runtimedoc=false
type ColumnValuer[T any] func(v Column) sqlfrag.Fragment

func AsValue[T any](v TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("?", v)
	}
}

func Value[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("?", v)
	}
}

func Incr[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? + ?", c, v)
	}
}

func Des[T any](v T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? - ?", c, v)
	}
}

func Eq[T comparable](expect T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? = ?", c, expect)
	}
}

func EqCol[T comparable](expect TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? = ?", c, expect)
	}
}

func NeqCol[T comparable](expect TypedColumn[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <> ?", c, expect)
	}
}

func In[T any](values ...T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		if len(values) == 0 {
			return nil
		}
		return sqlfrag.Pair("? IN (?)", c, values)
	}
}

func InSeq[T any](values iter.Seq[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IN (?)", c, xiter.Map(values, func(x T) any {
			return x
		}))
	}
}

func NotIn[T any](values ...T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		if len(values) == 0 {
			return nil
		}

		return sqlfrag.Pair("? NOT IN (?)", c, values)
	}
}

func NotInSeq[T any](values iter.Seq[T]) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT IN (?)", c, xiter.Map(values, func(x T) any {
			return x
		}))
	}
}

func IsNull[T any]() ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IS NULL", c)
	}
}

func IsNotNull[T any]() ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? IS NOT NULL", c)
	}
}

func Neq[T any](expect T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <> ?", c, expect)
	}
}

func Like[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, "%"+s+"%")
	}
}

func NotLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT LIKE ?", c, "%"+s+"%")
	}
}

func LeftLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, "%"+s)
	}
}

func RightLike[T ~string](s T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? LIKE ?", c, s+"%")
	}
}

func Between[T comparable](leftValue T, rightValue T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

func NotBetween[T comparable](leftValue T, rightValue T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? NOT BETWEEN ? AND ?", c, leftValue, rightValue)
	}
}

func Gt[T comparable](min T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? > ?", c, min)
	}
}

func Gte[T comparable](min T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? >= ?", c, min)
	}
}

func Lt[T comparable](max T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? < ?", c, max)
	}
}

func Lte[T comparable](max T) ColumnValuer[T] {
	return func(c Column) sqlfrag.Fragment {
		return sqlfrag.Pair("? <= ?", c, max)
	}
}
