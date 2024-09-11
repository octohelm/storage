package compose

import (
	"fmt"
	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func ApplyMutationFromFilter[M any, T comparable](m dal.Mutation[M], col sqlbuilder.TypedColumn[T], f *filter.Filter[T]) dal.Mutation[M] {
	if where, ok := WhereFromFilter(col, f); ok {
		return m.WhereAnd(where)
	}
	return m
}

func ApplyQuerierFromFilter[T comparable](q dal.Querier, col sqlbuilder.TypedColumn[T], f *filter.Filter[T]) dal.Querier {
	if q.ExistsTable(col.T()) {
		if where, ok := WhereFromFilter(col, f); ok {
			return q.WhereAnd(where)
		}
	}
	return q
}

func WhereFromFilter[T comparable](col sqlbuilder.TypedColumn[T], f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
	if f.IsZero() {
		return nil, false
	}

	switch f.Op() {
	case filter.OP__AND:
		rules := filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
			return WhereFromFilter(col, f)
		})
		return sqlbuilder.And(rules...), true
	case filter.OP__OR:
		return sqlbuilder.Or(
			filter.MapFilter(f.Args(), func(f *filter.Filter[T]) (sqlfrag.Fragment, bool) {
				return WhereFromFilter(col, f)
			})...,
		), true
	case filter.OP__NOTIN:
		return col.V(sqlbuilder.NotIn(
			filter.MapWhere(f.Args(), func(arg filter.Arg) (T, bool) {
				if r, ok := arg.(filter.Value[T]); ok {
					return r.Value(), true
				}
				return *new(T), false
			})...,
		)), true
	case filter.OP__IN:
		return col.V(sqlbuilder.In(
			filter.MapWhere(f.Args(), func(arg filter.Arg) (T, bool) {
				if r, ok := arg.(filter.Value[T]); ok {
					return r.Value(), true
				}
				return *new(T), false
			})...,
		)), true
	case filter.OP__EQ:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Eq(v)), true
	case filter.OP__NEQ:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Neq(v)), true
	case filter.OP__GT:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Gt(v)), true
	case filter.OP__GTE:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Gte(v)), true
	case filter.OP__LT:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Lt(v)), true
	case filter.OP__LTE:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}
		return col.V(sqlbuilder.Lte(v)), true
	case filter.OP__PREFIX:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}

		s := fmt.Sprintf("%v", v)

		return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
			return col.Fragment("# LIKE ?", s+"%")
		}), true
	case filter.OP__SUFFIX:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}

		s := fmt.Sprintf("%v", v)

		return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
			return col.Fragment("# LIKE ?", "%"+s)
		}), true

	case filter.OP__CONTAINS:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}

		s := fmt.Sprintf("%v", v)

		return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
			return col.Fragment("# LIKE ?", "%"+s+"%")
		}), true
	case filter.OP__NOTCONTAINS:
		v, ok := filter.First(f.Args(), func(arg filter.Arg) (T, bool) {
			if r, ok := arg.(filter.Value[T]); ok {
				return r.Value(), true
			}
			return *new(T), false
		})
		if !ok {
			return nil, false
		}

		s := fmt.Sprintf("%v", v)

		return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
			return col.Fragment("# NOT LIKE ?", "%"+s+"%")
		}), true
	default:

	}

	return nil, false
}
