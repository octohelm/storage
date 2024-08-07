package compose

import (
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func ApplyQuerierFromFilter[T comparable](q dal.Querier, col sqlbuilder.TypedColumn[T], f *filter.Filter[T]) dal.Querier {
	if q.ExistsTable(col.T()) {
		if where, ok := WhereFromFilter(col, f); ok {
			return q.WhereAnd(where)
		}
	}
	return q
}

func WhereFromFilter[T comparable](col sqlbuilder.TypedColumn[T], f *filter.Filter[T]) (sqlbuilder.SqlExpr, bool) {
	if f.IsZero() {
		return nil, false
	}

	switch f.Op() {
	case filter.OP__AND:
		rules := filter.MapFilter(f.Args(), func(arg filter.Arg) (sqlbuilder.SqlExpr, bool) {
			switch x := arg.(type) {
			case *filter.Filter[T]:
				return WhereFromFilter(col, x)
			case filter.Filter[T]:
				return WhereFromFilter(col, &x)
			}
			return nil, false
		})
		return sqlbuilder.And(rules...), true
	case filter.OP__OR:
		return sqlbuilder.Or(
			filter.MapFilter(f.Args(), func(arg filter.Arg) (sqlbuilder.SqlExpr, bool) {
				switch x := arg.(type) {
				case *filter.Filter[T]:
					return WhereFromFilter(col, x)
				case filter.Filter[T]:
					return WhereFromFilter(col, &x)
				}
				return nil, false
			})...,
		), true
	case filter.OP__NOTIN:
		return col.V(sqlbuilder.NotIn(
			filter.MapFilter(f.Args(), func(arg filter.Arg) (T, bool) {
				if r, ok := arg.(filter.Value[T]); ok {
					return r.Value(), true
				}
				return *new(T), false
			})...,
		)), true
	case filter.OP__IN:
		return col.V(sqlbuilder.In(
			filter.MapFilter(f.Args(), func(arg filter.Arg) (T, bool) {
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
	default:

	}

	return nil, false
}
