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

type Querier interface {
	sqlbuilder.SqlExpr

	Join(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier
	CrossJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier
	LeftJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier
	RightJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier
	FullJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier

	Where(where sqlbuilder.SqlCondition) Querier

	OrderBy(orders ...*sqlbuilder.Order) Querier

	GroupBy(cols ...sqlbuilder.SqlExpr) Querier
	Having(where sqlbuilder.SqlCondition) Querier

	Limit(v int64) Querier
	Offset(v int64) Querier

	Distinct() Querier
	Select(target sqlbuilder.SqlExpr) Querier

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
	having  sqlbuilder.SqlCondition

	limit  int64
	offset int64

	distinct bool

	where   sqlbuilder.SqlCondition
	project sqlbuilder.SqlExpr

	joins []sqlbuilder.Addition

	feature

	recv any
}

func (q querier) CrossJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier {
	q.joins = append(q.joins, sqlbuilder.CrossJoin(t).On(on))
	return &q
}

func (q querier) LeftJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier {
	q.joins = append(q.joins, sqlbuilder.LeftJoin(t).On(on))
	return &q
}

func (q querier) RightJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier {
	q.joins = append(q.joins, sqlbuilder.RightJoin(t).On(on))
	return &q
}

func (q querier) FullJoin(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier {
	q.joins = append(q.joins, sqlbuilder.FullJoin(t).On(on))
	return &q
}

func (q querier) Join(t sqlbuilder.Table, on sqlbuilder.SqlCondition) Querier {
	q.joins = append(q.joins, sqlbuilder.Join(t).On(on))
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
		tpe = tpe.Elem()
		if tpe.Kind() == reflect.Struct {
			return reflect.New(tpe).Interface().(sqlbuilder.Model)
		}
	}
	return nil
}

func (q querier) Scan(v any) Querier {
	if q.project == nil {
		if m, ok := resolveModel(v).(sqlbuilder.Model); ok {
			q.project = sqlbuilder.ColumnsByStruct(m)
		}
	}
	q.recv = v
	return &q
}

func (q querier) Select(project sqlbuilder.SqlExpr) Querier {
	q.project = project
	return &q
}

func (q querier) Where(where sqlbuilder.SqlCondition) Querier {
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

func (q querier) Having(having sqlbuilder.SqlCondition) Querier {
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

func (q *querier) buildWhere(t sqlbuilder.Table) sqlbuilder.SqlCondition {
	if q.feature.softDelete {
		if newModel, ok := q.from.(interface{ New() sqlbuilder.Model }); ok {
			m := newModel.New()
			if soft, ok := m.(ModelWithSoftDelete); ok {
				f, _ := soft.SoftDeleteFieldAndValue()
				return sqlbuilder.And(q.where, t.F(f).Eq(0))
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

	additions := make([]sqlbuilder.Addition, 0)

	if where := q.buildWhere(from); where != nil {
		additions = append(additions, sqlbuilder.Where(where))
	}

	if n := len(q.joins); n > 0 {
		additions = append(additions, q.joins...)
	}

	if n := len(q.orders); n > 0 {
		additions = append(additions, sqlbuilder.OrderBy(q.orders...))
	}

	if n := len(q.groupBy); n > 0 {
		additions = append(additions, sqlbuilder.GroupBy(q.groupBy...).Having(q.having))
	}

	if q.limit > 0 {
		additions = append(additions, sqlbuilder.Limit(q.limit).Offset(q.offset))
	}

	return sqlbuilder.Select(q.project).From(from, additions...)
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
