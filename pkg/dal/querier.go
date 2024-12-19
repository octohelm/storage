package dal

import (
	"context"
	"errors"
	"fmt"
	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"iter"
	"reflect"
)

// Intersect(q Querier) Querier
// Except(q Querier) Querier
func InSelect[T any](col sqlbuilder.TypedColumn[T], q Querier) sqlbuilder.ColumnValuer[T] {
	return func(v sqlbuilder.Column) sqlfrag.Fragment {
		ex := q.Select(col)
		if ex.IsNil() {
			return nil
		}
		return sqlfrag.Pair("? IN (?)", v, ex)
	}
}

func NotInSelect[T any](col sqlbuilder.TypedColumn[T], q Querier) sqlbuilder.ColumnValuer[T] {
	return func(v sqlbuilder.Column) sqlfrag.Fragment {
		ex := q.Select(col)
		if ex.IsNil() {
			return nil
		}
		return sqlfrag.Pair("? NOT IN (?)", v, ex)
	}
}

type QuerierPatcher interface {
	ApplyQuerier(q Querier) Querier
}

type Querier interface {
	sqlfrag.Fragment

	ExistsTable(table sqlbuilder.Table) bool
	Apply(patchers ...QuerierPatcher) Querier

	AsSelect() sqlbuilder.SelectStatement

	With(t sqlbuilder.Table, build sqlbuilder.BuildSubQuery, modifiers ...string) Querier
	AsTemporaryTable(tableName string) TemporaryTable

	Join(t sqlbuilder.Table, on sqlfrag.Fragment) Querier
	CrossJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier
	LeftJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier
	RightJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier
	FullJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier

	Where(where sqlfrag.Fragment) Querier
	WhereAnd(where sqlfrag.Fragment) Querier
	WhereOr(where sqlfrag.Fragment) Querier

	OrderBy(orders ...sqlbuilder.Order) Querier

	GroupBy(cols ...sqlfrag.Fragment) Querier
	Having(where sqlfrag.Fragment) Querier

	Limit(v int64) Querier
	Offset(v int64) Querier

	Distinct(extras ...sqlfrag.Fragment) Querier
	Select(projects ...sqlfrag.Fragment) Querier

	Scan(v any) Querier

	Find(ctx context.Context) error
	Count(ctx context.Context) (int, error)
}

func From(from sqlbuilder.Table, fns ...OptionFunc) Querier {
	q := &querier{
		from:   from,
		tables: []sqlbuilder.Table{from},
		limit:  -1,
		feature: feature{
			softDelete: true,
		},
	}

	for i := range fns {
		fns[i](q)
	}

	if tmpT, ok := from.(QuerierPatcher); ok {
		return q.Apply(tmpT)
	}

	return q
}

type querier struct {
	from   sqlbuilder.Table
	tables []sqlbuilder.Table

	withStmt *sqlbuilder.WithStmt

	orders []sqlbuilder.Order

	distinct []sqlfrag.Fragment
	groupBy  []sqlfrag.Fragment
	having   sqlfrag.Fragment

	limit  int64
	offset int64

	where    sqlfrag.Fragment
	projects []sqlfrag.Fragment

	joins []sqlbuilder.Addition

	feature

	recv any
}

func (q *querier) ExistsTable(table sqlbuilder.Table) bool {
	for _, t := range q.tables {
		if t == table || t.TableName() == table.TableName() {
			return true
		}
	}
	return false
}

func (q *querier) Apply(patchers ...QuerierPatcher) Querier {
	var applied Querier = q

	for _, p := range patchers {
		if p != nil {
			applied = p.ApplyQuerier(applied)
		}
	}

	return applied
}

func (q querier) With(t sqlbuilder.Table, build sqlbuilder.BuildSubQuery, modifiers ...string) Querier {
	q.tables = append(q.tables, t)
	if q.withStmt == nil {
		q.withStmt = sqlbuilder.With(t, build, modifiers...)
		return &q
	}
	q.withStmt = q.withStmt.With(t, build)
	return &q
}

func (q querier) CrossJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier {
	q.tables = append(q.tables, t)
	q.joins = append(q.joins, sqlbuilder.CrossJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) LeftJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier {
	q.tables = append(q.tables, t)
	q.joins = append(q.joins, sqlbuilder.LeftJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) RightJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier {
	q.tables = append(q.tables, t)
	q.joins = append(q.joins, sqlbuilder.RightJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) FullJoin(t sqlbuilder.Table, on sqlfrag.Fragment) Querier {
	q.tables = append(q.tables, t)
	q.joins = append(q.joins, sqlbuilder.FullJoin(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q querier) Join(t sqlbuilder.Table, on sqlfrag.Fragment) Querier {
	q.tables = append(q.tables, t)
	q.joins = append(q.joins, sqlbuilder.Join(t).On(sqlbuilder.AsCond(on)))
	return &q
}

func (q *querier) IsNil() bool {
	if q.whereStmtNotEmpty {
		return sqlfrag.IsNil(q.where) || q.from == nil
	}
	return q.from == nil
}

func (q *querier) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return q.build().Frag(ctx)
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
			q.projects = []sqlfrag.Fragment{columnsByStruct(m)}
		}
	}
	q.recv = v
	return &q
}

func columnsByStruct(v any) sqlfrag.Fragment {
	return sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
		return func(yield func(string, []any) bool) {
			i := 0

			for fieldValue := range structs.AllFieldValue(ctx, v) {
				if i > 0 {
					if !yield(",", nil) {
						return
					}
				}

				if fieldValue.TableName != "" {
					if !yield(fmt.Sprintf("%s.%s AS %s", fieldValue.TableName, fieldValue.Field.Name, sqlfrag.SafeProjected(fieldValue.TableName, fieldValue.Field.Name)), nil) {
						return
					}
				} else {
					if !yield(fieldValue.Field.Name, nil) {
						return
					}
				}

				i++
			}
		}
	})
}

func (q querier) Select(projects ...sqlfrag.Fragment) Querier {
	q.projects = projects
	return &q
}

func (q querier) Where(where sqlfrag.Fragment) Querier {
	q.where = where
	return &q
}

func (q querier) WhereAnd(where sqlfrag.Fragment) Querier {
	q.where = sqlbuilder.And(q.where, where)
	return &q
}

func (q querier) WhereOr(where sqlfrag.Fragment) Querier {
	q.where = sqlbuilder.Or(q.where, where)
	return &q
}

func (q querier) OrderBy(orders ...sqlbuilder.Order) Querier {
	q.orders = orders
	return &q
}

func (q querier) GroupBy(cols ...sqlfrag.Fragment) Querier {
	q.groupBy = cols
	return &q
}

func (q querier) Having(having sqlfrag.Fragment) Querier {
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

func (q querier) Distinct(extras ...sqlfrag.Fragment) Querier {
	q.distinct = extras
	return &q
}

func (q *querier) buildWhere(t sqlbuilder.Table) sqlfrag.Fragment {
	if q.feature.softDelete {
		if newModel, ok := q.from.(ModelNewer); ok {
			m := newModel.New()

			if soft, ok := m.(ModelWithSoftDelete); ok {
				f, _ := soft.SoftDeleteFieldAndZeroValue()

				return sqlbuilder.And(
					q.where,
					sqlbuilder.CastColumn[int](t.F(f)).V(sqlbuilder.Eq(0)),
				)
			}
		}
	}
	return q.where
}

func (q *querier) AsSelect() sqlbuilder.SelectStatement {
	from := q.from

	modifies := make([]sqlfrag.Fragment, 0)

	if q.distinct != nil {
		modifies = append(modifies, sqlfrag.Const("DISTINCT"))

		if len(q.distinct) > 0 {
			modifies = append(modifies, q.distinct...)
		}
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

	var projects sqlfrag.Fragment

	if len(q.projects) > 0 {
		projects = sqlbuilder.MultiMayAutoAlias(q.projects...)
	}

	return sqlbuilder.Select(projects, modifies...).From(from, additions...)
}

func (q *querier) build() sqlfrag.Fragment {
	if q.withStmt != nil {
		return q.withStmt.Exec(func(tables ...sqlbuilder.Table) sqlfrag.Fragment {
			return q.AsSelect()
		})
	}
	return q.AsSelect()
}

func (q *querier) Count(ctx context.Context) (int, error) {
	var c int
	if err := q.Limit(-1).Select(sqlbuilder.Count()).Scan(&c).Find(ctx); err != nil {
		return 0, err
	}
	return c, nil
}

func (q *querier) Find(ctx context.Context) error {
	s := SessionFor(ctx, q.from)

	if q.recv == nil {
		return errors.New("missing receiver. need to use Scan to bind one")
	}
	rows, err := s.Adapter().Query(ctx, q.build())
	if err != nil {
		return err
	}

	done := make(chan error)

	go func() {
		defer close(done)

		if err := scanner.Scan(ctx, rows, q.recv); err != nil {
			if errors.Is(err, ErrSkipScan) || errors.Is(err, context.Canceled) {
				done <- nil
				return
			}
			done <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	default:
		return <-done
	}
}

type ScanIterator = scanner.ScanIterator

var ErrSkipScan = errors.New("scan skip")

type TemporaryTable interface {
	sqlbuilder.Table
	session.TableWrapper
	QuerierPatcher
}

func (q *querier) AsTemporaryTable(tableName string) TemporaryTable {
	projects := q.projects

	cols := make([]sqlfrag.Fragment, 0, len(projects))

	for _, p := range projects {
		if col, ok := p.(sqlbuilder.Column); ok {
			cols = append(cols, col)
		}
	}

	tmpT := sqlbuilder.T(tableName, cols...)

	return &tmpTable{
		Table:  tmpT,
		origin: q.from,
		build: func(table sqlbuilder.Table) sqlfrag.Fragment {
			return q
		},
	}
}

type tmpTable struct {
	sqlbuilder.Table
	origin sqlbuilder.Table
	build  func(table sqlbuilder.Table) sqlfrag.Fragment
}

func (t *tmpTable) Unwrap() sqlbuilder.Model {
	return t.origin
}

func (t *tmpTable) ApplyQuerier(q Querier) Querier {
	return q.With(t.Table, t.build)
}
