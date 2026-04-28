package filter

import (
	"fmt"
	"iter"

	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe"
)

// AsWhere 把 filter.Filter 转为 sqlpipe 的 WHERE 操作符。
func AsWhere[M sqlpipe.Model, T comparable](col modelscoped.TypedColumn[M, T], f *filter.Filter[T]) sqlpipe.SourceOperator[M] {
	return sqlpipe.NewWhere(sqlpipe.FilterOpAnd, col, func(v sqlbuilder.Column) sqlfrag.Fragment {
		return BuildWhere(f, func(op filter.Op, seq iter.Seq[T], create func(seq iter.Seq[T]) sqlbuilder.ColumnValuer[T]) sqlfrag.Fragment {
			return col.V(create(seq))
		})
	})
}

// BuildWhere 把过滤规则树构造成 SQL 条件片段。
func BuildWhere[T comparable](f *filter.Filter[T], apply func(op filter.Op, seq iter.Seq[T], create func(seq iter.Seq[T]) sqlbuilder.ColumnValuer[T]) sqlfrag.Fragment) sqlfrag.Fragment {
	if f == nil || f.IsZero() {
		return nil
	}

	op := f.Op()
	switch op {
	case filter.OP__AND:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
			return BuildWhere[T](f, apply), true
		})
		return sqlbuilder.AndSeq(rules)
	case filter.OP__OR:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
			return BuildWhere[T](f, apply), true
		})
		return sqlbuilder.OrSeq(rules)
	default:
		return apply(op, Values(f), func(seq iter.Seq[T]) sqlbuilder.ColumnValuer[T] {
			switch op {
			case filter.OP__IN:
				return sqlbuilder.InSeq(seq)
			case filter.OP__NOTIN:
				return sqlbuilder.NotInSeq(seq)
			default:
				for v := range seq {
					switch op {
					case filter.OP__EQ:
						return sqlbuilder.Eq(v)
					case filter.OP__NEQ:
						return sqlbuilder.Neq(v)
					case filter.OP__GT:
						return sqlbuilder.Gt(v)
					case filter.OP__GTE:
						return sqlbuilder.Gte(v)
					case filter.OP__LT:
						return sqlbuilder.Lt(v)
					case filter.OP__LTE:
						return sqlbuilder.Lte(v)
					case filter.OP__PREFIX:
						return func(col sqlbuilder.Column) sqlfrag.Fragment {
							return col.Fragment("# LIKE ?", fmt.Sprintf("%v", v)+"%")
						}
					case filter.OP__SUFFIX:
						return func(col sqlbuilder.Column) sqlfrag.Fragment {
							return col.Fragment("# LIKE ?", "%"+fmt.Sprintf("%v", v))
						}
					case filter.OP__CONTAINS:
						return func(col sqlbuilder.Column) sqlfrag.Fragment {
							return col.Fragment("# LIKE ?", "%"+fmt.Sprintf("%v", v)+"%")
						}
					case filter.OP__NOTCONTAINS:
						return func(col sqlbuilder.Column) sqlfrag.Fragment {
							return col.Fragment("# NOT LIKE ?", "%"+fmt.Sprintf("%v", v)+"%")
						}
					default:

					}
				}
			}

			return nil
		})
	}
}

// SubFilters 返回复合过滤器中的子过滤器。
func SubFilters[T comparable](f *filter.Filter[T]) iter.Seq[*filter.Filter[T]] {
	return func(yield func(*filter.Filter[T]) bool) {
		for arg := range f.Args() {
			switch x := any(arg).(type) {
			case filter.Filter[T]:
				if !yield(&x) {
					return
				}
			case *filter.Filter[T]:
				if !yield(x) {
					return
				}
			}
		}
	}
}

// Values 返回过滤器中的字面量值序列。
func Values[T comparable](f *filter.Filter[T]) iter.Seq[T] {
	return func(yield func(value T) bool) {
		for arg := range f.Args() {
			switch x := any(arg).(type) {
			case filter.Value[T]:
				if !yield(x.Value()) {
					return
				}
			}
		}
	}
}
