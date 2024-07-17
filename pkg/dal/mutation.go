package dal

import (
	"context"
	"database/sql/driver"
	"github.com/octohelm/x/slices"
	"time"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/datatypes"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Prepare[T any](v *T) Mutation[T] {
	if m, ok := any(v).(ModelWithCreationTime); ok {
		m.MarkCreatedAt()
	}

	return &mutation[T]{
		target: v,
		feature: feature{
			softDelete: true,
		},
	}
}

type Mutation[T any] interface {
	IncludesZero(zeroFields ...sqlbuilder.Column) Mutation[T]

	FromSelect(q Querier, cols ...sqlbuilder.Column) Mutation[T]

	ForDelete(opts ...OptionFunc) Mutation[T]
	ForUpdateSet(assignments ...sqlbuilder.Assignment) Mutation[T]

	Where(where sqlbuilder.SqlExpr) Mutation[T]

	OnConflict(cols sqlbuilder.ColumnCollection) Mutation[T]
	DoNothing() Mutation[T]
	DoUpdateSet(cols ...sqlbuilder.Column) Mutation[T]
	DoWith(func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition) Mutation[T]

	Returning(cols ...sqlbuilder.SqlExpr) Mutation[T]
	Scan(recv any) Mutation[T]

	Save(ctx context.Context) error
}

type mutation[T any] struct {
	target             *T
	recv               any
	zeroFieldsIncludes []sqlbuilder.Column

	assignmentsForUpdate sqlbuilder.Assignments
	where                sqlbuilder.SqlExpr

	conflict              sqlbuilder.ColumnCollection
	onConflictDoWith      func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition
	onConflictDoUpdateSet []sqlbuilder.Column

	fromSelect *fromSelect

	returning []sqlbuilder.SqlExpr

	forDelete bool

	feature
}

type fromSelect struct {
	columns []sqlbuilder.Column
	values  sqlbuilder.SelectStatement
}

type DeleteFunc func()

func (c mutation[T]) IncludesZero(zeroFields ...sqlbuilder.Column) Mutation[T] {
	c.zeroFieldsIncludes = zeroFields
	return &c
}

func (c mutation[T]) FromSelect(q Querier, cols ...sqlbuilder.Column) Mutation[T] {
	c.fromSelect = &fromSelect{
		columns: cols,
		values:  q.AsSelect(),
	}
	return &c
}

func (c mutation[T]) ForDelete(fns ...OptionFunc) Mutation[T] {
	c.forDelete = true
	for i := range fns {
		fns[i](&c)
	}
	return &c
}

func (c mutation[T]) ForUpdateSet(assignments ...sqlbuilder.Assignment) Mutation[T] {
	c.assignmentsForUpdate = assignments
	return &c
}

func (c mutation[T]) Where(where sqlbuilder.SqlExpr) Mutation[T] {
	c.where = where
	return &c
}

func (c mutation[T]) OnConflict(cols sqlbuilder.ColumnCollection) Mutation[T] {
	c.conflict = cols
	return &c
}

func (c mutation[T]) DoNothing() Mutation[T] {
	c.onConflictDoUpdateSet = nil
	return &c
}

func (c mutation[T]) DoWith(fn func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition) Mutation[T] {
	c.onConflictDoWith = fn
	return &c
}

func (c mutation[T]) DoUpdateSet(cols ...sqlbuilder.Column) Mutation[T] {
	c.onConflictDoUpdateSet = cols
	return &c
}

func (c mutation[T]) Returning(cols ...sqlbuilder.SqlExpr) Mutation[T] {
	if len(cols) != 0 {
		c.returning = cols
	} else {
		c.returning = make([]sqlbuilder.SqlExpr, 0)
	}
	return &c
}

func (c mutation[T]) Scan(recv any) Mutation[T] {
	c.recv = recv
	return &c
}

func (c *mutation[T]) Save(ctx context.Context) error {
	s := SessionFor(ctx, c.target)
	if c.forDelete {
		return c.del(ctx, s.T(c.target), s)
	}
	return c.insertOrUpdate(ctx, s.T(c.target), s)
}

func (c *mutation[T]) buildWhere(t sqlbuilder.Table) sqlbuilder.SqlCondition {
	if c.where == nil {
		return nil
	}
	where := c.where
	if c.feature.softDelete {
		if soft, ok := any(c.target).(ModelWithSoftDelete); ok {
			f, notDeletedValue := soft.SoftDeleteFieldAndZeroValue()
			return sqlbuilder.And(
				where,
				t.F(f).Expr("# = ?", notDeletedValue),
			)
		}
	}
	return sqlbuilder.AsCond(where)
}

func (c *mutation[T]) del(ctx context.Context, t sqlbuilder.Table, s Session) error {
	where := c.buildWhere(t)
	if where == nil {
		// never delete without condition
		return nil
	}

	var stmt sqlbuilder.SqlExpr

	additions, hasReturning := c.withReturning(t, nil)

	if c.feature.softDelete {
		if soft, ok := any(c.target).(ModelWithSoftDelete); ok {
			if x, ok := any(c.target).(DeletedAtMarker); ok {
				x.MarkDeletedAt()
			}

			f, _ := soft.SoftDeleteFieldAndZeroValue()

			var softDeleteValue driver.Value
			if v, ok := ctx.(SoftDeleteValueGetter); ok {
				softDeleteValue = v.GetDeletedAt()
			} else {
				softDeleteValue = datatypes.Timestamp(time.Now())
			}

			col := t.F(f)
			stmt = sqlbuilder.Update(t).Where(where, additions...).Set(
				sqlbuilder.ColumnsAndValues(col, softDeleteValue),
			)
		}
	}

	if stmt == nil {
		stmt = sqlbuilder.Delete().From(t, append([]sqlbuilder.Addition{sqlbuilder.Where(where)}, additions...)...)
	}

	return c.exec(ctx, s, stmt, hasReturning)
}

func (c *mutation[T]) insertOrUpdate(ctx context.Context, t sqlbuilder.Table, s Session) error {
	additions := make([]sqlbuilder.Addition, 0)

	if c.conflict != nil && c.conflict.Len() > 0 {
		onConflict := sqlbuilder.OnConflict(c.conflict)

		if onConflictDoWith := c.onConflictDoWith; onConflictDoWith != nil {
			additions = append(additions, onConflictDoWith(onConflict))
		} else {
			cols := c.onConflictDoUpdateSet
			if cols == nil {
				// FIXME ugly hack
				// sqlite will not RETURNING when ON CONFLICT DO NOTHING
				c.conflict.RangeCol(func(col sqlbuilder.Column, idx int) bool {
					cols = append(cols, col)
					return true
				})
			}

			assignments := make([]sqlbuilder.Assignment, len(cols))

			for idx, col := range cols {
				assignments[idx] = sqlbuilder.ColumnsAndValues(
					col, col.Expr("EXCLUDED.?", sqlbuilder.Expr(col.Name())),
				)
			}

			onConflict = onConflict.DoUpdateSet(assignments...)
			additions = append(additions, onConflict)
		}
	}

	additions, hasReturning := c.withReturning(t, additions)

	zeroFieldsIncludes := make([]string, len(c.zeroFieldsIncludes))

	for i := range zeroFieldsIncludes {
		zeroFieldsIncludes[i] = c.zeroFieldsIncludes[i].FieldName()
	}

	fieldValues := sqlbuilder.FieldValuesFromStructByNonZero(c.target, zeroFieldsIncludes...)

	var stmt sqlbuilder.SqlExpr

	if where := c.buildWhere(t); where != nil {
		assignmentsForUpdate := c.assignmentsForUpdate
		if len(assignmentsForUpdate) == 0 {
			assignmentsForUpdate = sqlbuilder.AssignmentsByFieldValues(t, fieldValues)
		}
		stmt = sqlbuilder.Update(t).
			Where(where, additions...).
			Set(assignmentsForUpdate...)
	} else if c.fromSelect != nil {
		stmt = sqlbuilder.Insert().Into(t, additions...).
			Values(
				t.Cols(slices.Map(c.fromSelect.columns, func(e sqlbuilder.Column) string {
					return e.Name()
				})...),
				c.fromSelect.values,
			)
	} else {
		cols, vals := sqlbuilder.ColumnsAndValuesByFieldValues(t, fieldValues)
		stmt = sqlbuilder.Insert().Into(t, additions...).
			Values(cols, vals...)
	}

	return c.exec(ctx, s, stmt, hasReturning)
}

func (c *mutation[T]) exec(ctx context.Context, s Session, stmt sqlbuilder.SqlExpr, hasReturning bool) error {
	if hasReturning {
		rows, err := s.Adapter().Query(ctx, stmt)
		if err != nil {
			return err
		}
		return scanner.Scan(ctx, rows, c.recv)
	}
	_, err := s.Adapter().Exec(ctx, stmt)
	return err
}

func (c *mutation[T]) withReturning(t sqlbuilder.Table, additions []sqlbuilder.Addition) ([]sqlbuilder.Addition, bool) {
	hasReturning := false

	if c.returning != nil {
		hasReturning = true

		if len(c.returning) == 0 {
			additions = append(additions, sqlbuilder.Returning(sqlbuilder.Expr("*")))
		} else {
			additions = append(additions, sqlbuilder.Returning(sqlbuilder.MultiMayAutoAlias(c.returning...)))
		}
	}

	return additions, hasReturning
}
