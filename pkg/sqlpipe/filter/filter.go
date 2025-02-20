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

func AsWhere[M sqlpipe.Model, T comparable](col modelscoped.TypedColumn[M, T], f *filter.Filter[T]) sqlpipe.SourceOperator[M] {
	return sqlpipe.NewWhere(sqlpipe.FilterOpAnd, col, func(v sqlbuilder.Column) sqlfrag.Fragment {
		return asWhere(col, f)
	})
}

func asWhere[M sqlpipe.Model, T comparable](col modelscoped.TypedColumn[M, T], f *filter.Filter[T]) sqlfrag.Fragment {
	if f == nil || f.IsZero() {
		return nil
	}
	switch f.Op() {
	case filter.OP__AND:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
			return asWhere(col, f), true
		})
		return sqlbuilder.AndSeq(rules)
	case filter.OP__OR:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
			return asWhere(col, f), true
		})
		return sqlbuilder.OrSeq(rules)
	case filter.OP__NOTIN:
		return col.V(sqlbuilder.NotInSeq(Values(f)))
	case filter.OP__IN:
		return col.V(sqlbuilder.InSeq(Values(f)))
	case filter.OP__EQ:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Eq(v))
		}
	case filter.OP__NEQ:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Neq(v))
		}
	case filter.OP__GT:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Gt(v))
		}
	case filter.OP__GTE:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Gte(v))
		}
	case filter.OP__LT:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Lt(v))
		}
	case filter.OP__LTE:
		if v, ok := pickValue(f); ok {
			return col.V(sqlbuilder.Lte(v))
		}
	case filter.OP__PREFIX:
		if v, ok := pickValue(f); ok {
			return col.Fragment("# LIKE ?", fmt.Sprintf("%v", v)+"%")
		}
	case filter.OP__SUFFIX:
		if v, ok := pickValue(f); ok {
			return col.Fragment("# LIKE ?", "%"+fmt.Sprintf("%v", v))
		}
	case filter.OP__CONTAINS:
		if v, ok := pickValue(f); ok {
			return col.Fragment("# LIKE ?", "%"+fmt.Sprintf("%v", v)+"%")
		}
	case filter.OP__NOTCONTAINS:
		if v, ok := pickValue(f); ok {
			return col.Fragment("# NOT LIKE ?", "%"+fmt.Sprintf("%v", v)+"%")
		}
	default:

	}

	return nil
}

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

func pickValue[T comparable](f *filter.Filter[T]) (T, bool) {
	for v := range Values(f) {
		return v, true
	}
	return *new(T), false
}
