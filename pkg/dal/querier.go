package dal

import (
	"context"
	"reflect"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/pkg/errors"
)

//Intersect(q Querier) Querier
//Except(q Querier) Querier

func InSelect[T any](col sqlbuilder.TypedColumn[T], q Querier) sqlbuilder.ColumnValueExpr[T] {
	return func(v sqlbuilder.Column) sqlbuilder.SqlExpr {
		return sqlbuilder.Expr("? IN (?)", v, q.Select(col))
	}
}

type Querier interface {
	sqlbuilder.SqlExpr

	Join(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier
	CrossJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier
	LeftJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier
	RightJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier
	FullJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier

	Where(where sqlbuilder.SqlExpr) Querier

	OrderBy(orders ...*sqlbuilder.Order) Querier

	GroupBy(cols ...sqlbuilder.SqlExpr) Querier
	Having(where sqlbuilder.SqlExpr) Querier

	Limit(v int64) Querier
	Offset(v int64) Querier

	Distinct() Querier
	Select(projects ...sqlbuilder.SqlExpr) Querier

	Scan(v any) Querier

	Find(ctx context.Context, s Session) error
	Count(ctx context.Context, s Session) (int, error)
}

func From(from sqlbuilder.Table, fns ...OptionFunc) Querier {
	q := &querier{
		from:  from,
		limit: -1,
		feature: feature{
			softDelete: true,
		},
	}

	for i := range fns {
		fns[i](q)
	}

	return q
}

type querier struct {
	from sqlbuilder.Table

	orders []*sqlbuilder.Order

	groupBy []sqlbuilder.SqlExpr
	having  sqlbuilder.SqlExpr

	limit  int64
	offset int64

	distinct bool

	where    sqlbuilder.SqlExpr
	projects []sqlbuilder.SqlExpr

	joins []sqlbuilder.Addition

	feature

	recv any
}

func (q querier) CrossJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier {
	q.joins = append(q.joins, sqlbuilder.CrossJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) LeftJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier {
	q.joins = append(q.joins, sqlbuilder.LeftJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) RightJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier {
	q.joins = append(q.joins, sqlbuilder.RightJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) FullJoin(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier {
	q.joins = append(q.joins, sqlbuilder.FullJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) Join(t sqlbuilder.Table, on sqlbuilder.SqlExpr) Querier {
	q.joins = append(q.joins, sqlbuilder.Join(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q *querier) IsNil() bool {
	return q.from == nil
}

func (q *querier) Ex(ctx context.Context) *sqlbuilder.Ex {
	return q.build().Ex(ctx)
}

func resolveModel(v any) any {
	if canNew, ok := v.(interface{ New() any }); ok {
		return canNew.New()
	} else {
		tpe := reflect.TypeOf(v)
		for tpe.Kind() == reflect.Ptr {
			tpe = tpe.Elem()
		}
		if tpe.Kind() == reflect.Struct {
			return reflect.New(tpe).Interface().(sqlbuilder.Model)
		}
	}
	return nil
}

func (q querier) Scan(v any) Querier {
	if len(q.projects) == 0 {
		if m, ok := resolveModel(v).(sqlbuilder.Model); ok {
			q.projects = []sqlbuilder.SqlExpr{sqlbuilder.ColumnsByStruct(m)}
		}
	}
	q.recv = v
	return &q
}

func (q querier) Select(projects ...sqlbuilder.SqlExpr) Querier {
	q.projects = projects
	return &q
}

func (q querier) Where(where sqlbuilder.SqlExpr) Querier {
	q.where = where
	return &q
}

func (q querier) OrderBy(orders ...*sqlbuilder.Order) Querier {
	q.orders = orders
	return &q
}

func (q querier) GroupBy(cols ...sqlbuilder.SqlExpr) Querier {
	q.groupBy = cols
	return &q
}

func (q querier) Having(having sqlbuilder.SqlExpr) Querier {
	q.having = having
	return &q
}

func (q querier) Limit(v int64) Querier {
	q.limit = v
	return &q
}

func (q querier) Offset(v int64) Querier {
	q.offset = v
	return &q
}

func (q querier) Distinct() Querier {
	q.distinct = true
	return &q
}

func (q *querier) buildWhere(t sqlbuilder.Table) sqlbuilder.SqlExpr {
	if q.feature.softDelete {
		if newModel, ok := q.from.(interface{ New() sqlbuilder.Model }); ok {
			m := newModel.New()
			if soft, ok := m.(ModelWithSoftDelete); ok {
				f, _ := soft.SoftDeleteFieldAndValue()
				return sqlbuilder.And(
					q.where,
					sqlbuilder.CastCol[int](t.F(f)).V(sqlbuilder.Eq(0)),
				)
			}
		}
	}
	return q.where
}

func (q *querier) build() sqlbuilder.SqlExpr {
	from := q.from

	modifies := make([]string, 0)

	if q.distinct {
		modifies = append(modifies, "DISTINCT")
	}

	additions := make([]sqlbuilder.Addition, 0, 10)

	if where := q.buildWhere(from); where != nil {
		additions = append(additions, sqlbuilder.Where(sqlbuilder.AsCond(where)))
	}

	if n := len(q.joins); n > 0 {
		additions = append(additions, q.joins...)
	}

	if n := len(q.orders); n > 0 {
		additions = append(additions, sqlbuilder.OrderBy(q.orders...))
	}

	if n := len(q.groupBy); n > 0 {
		additions = append(additions, sqlbuilder.GroupBy(q.groupBy...).Having(sqlbuilder.AsCond(q.having)))
	}

	if q.limit > 0 {
		additions = append(additions, sqlbuilder.Limit(q.limit).Offset(q.offset))
	}

	if q.projects == nil {
		return sqlbuilder.Select(nil).From(from, additions...)
	}

	return sqlbuilder.Select(sqlbuilder.MultiMayAutoAlias(q.projects...)).From(from, additions...)
}

func (q *querier) Find(ctx context.Context, s Session) error {
	if q.recv == nil {
		return errors.New("missing receiver. need to use Scan to bind one")
	}
	rows, err := s.Adapter().Query(ctx, q.build())
	if err != nil {
		return err
	}
	if err := scanner.Scan(ctx, rows, q.recv); err != nil {
		return err
	}
	return err
}

func (q *querier) Count(ctx context.Context, s Session) (int, error) {
	var c int
	if err := q.Limit(-1).Select(sqlbuilder.Count()).Scan(&c).Find(ctx, s); err != nil {
		return 0, err
	}
	return c, nil
}
