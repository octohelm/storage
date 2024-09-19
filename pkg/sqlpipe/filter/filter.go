package filter

import (
	"fmt"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe"
)

func AsWhere[M sqlpipe.Model, T comparable](col modelscoped.TypedColumn[M, T], f *filter.Filter[T]) sqlpipe.SourceOperator[M] {
	return asWhere(sqlpipe.FilterOpAnd, col, f)
}

func asWhere[M sqlpipe.Model, T comparable](op sqlpipe.FilterOp, col modelscoped.TypedColumn[M, T], f *filter.Filter[T]) sqlpipe.SourceOperator[M] {
	if f == nil || f.IsZero() {
		return sqlpipe.NewWhere(op, col, func(v sqlbuilder.Column) sqlfrag.Fragment {
			return nil
		})
	}

	switch f.Op() {
	case filter.OP__AND:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlpipe.SourceOperator[M], bool) {
			return asWhere[M, T](sqlpipe.FilterOpAnd, col, f), true
		})

		return sqlpipe.SourceOperatorFunc[M](sqlpipe.OperatorFilter, func(src sqlpipe.Source[M]) sqlpipe.Source[M] {
			return src.Pipe(rules...)
		})
	case filter.OP__OR:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlpipe.SourceOperator[M], bool) {
			return asWhere[M, T](sqlpipe.FilterOpOr, col, f), true
		})
		return sqlpipe.SourceOperatorFunc[M](sqlpipe.OperatorFilter, func(src sqlpipe.Source[M]) sqlpipe.Source[M] {
			return src.Pipe(rules...)
		})
	case filter.OP__NOTIN:
		return sqlpipe.NewWhere(op, col, sqlbuilder.NotIn(pickValues(f)...))
	case filter.OP__IN:
		return sqlpipe.NewWhere(op, col, sqlbuilder.In(pickValues(f)...))
	case filter.OP__EQ:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Eq(v))
		}
	case filter.OP__NEQ:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Neq(v))
		}
	case filter.OP__GT:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Gt(v))
		}
	case filter.OP__GTE:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Gte(v))
		}
	case filter.OP__LT:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Lt(v))
		}
	case filter.OP__LTE:
		if v, ok := pickValue(f); ok {
			return sqlpipe.NewWhere(op, col, sqlbuilder.Lte(v))
		}
	case filter.OP__PREFIX:
		if v, ok := pickValue(f); ok {
			s := fmt.Sprintf("%v", v)

			return sqlpipe.NewWhere(op, col, func(col sqlbuilder.Column) sqlfrag.Fragment {
				return col.Fragment("# LIKE ?", s+"%")
			})
		}
	case filter.OP__SUFFIX:
		if v, ok := pickValue(f); ok {
			s := fmt.Sprintf("%v", v)

			return sqlpipe.NewWhere(op, col, func(col sqlbuilder.Column) sqlfrag.Fragment {
				return col.Fragment("# LIKE ?", "%"+s)
			})
		}
	case filter.OP__CONTAINS:
		if v, ok := pickValue(f); ok {
			s := fmt.Sprintf("%v", v)

			return sqlpipe.NewWhere(op, col, func(col sqlbuilder.Column) sqlfrag.Fragment {
				return col.Fragment("# LIKE ?", "%"+s+"%")
			})
		}
	case filter.OP__NOTCONTAINS:
		if v, ok := pickValue(f); ok {
			s := fmt.Sprintf("%v", v)

			return sqlpipe.NewWhere(op, col, func(col sqlbuilder.Column) sqlfrag.Fragment {
				return col.Fragment("# NOT LIKE ?", "%"+s+"%")
			})
		}
	default:

	}

	return sqlpipe.NewWhere(op, col, func(v sqlbuilder.Column) sqlfrag.Fragment {
		return nil
	})
}

func SubFilters[T comparable](f *filter.Filter[T]) iter.Seq[*filter.Filter[T]] {
	return func(yield func(*filter.Filter[T]) bool) {
		for _, arg := range f.Args() {
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

func Values[T comparable](f *filter.Filter[T]) iter.Seq[T] {
	return func(yield func(value T) bool) {
		for _, arg := range f.Args() {
			switch x := any(arg).(type) {
			case filter.Value[T]:
				if !yield(x.Value()) {
					return
				}
			}
		}
	}
}

func pickValues[T comparable](f *filter.Filter[T]) []T {
	return slices.Collect(Values(f))
}

func pickValue[T comparable](f *filter.Filter[T]) (T, bool) {
	for v := range Values(f) {
		return v, true
	}
	return *new(T), false
}
