package dal

import (
	"context"
	"database/sql/driver"
	"fmt"
	"iter"
	"slices"
	"time"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/datatypes"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqlfrag"
	slicesx "github.com/octohelm/x/slices"
)

// Prepare
// Deprecated
// use Insert InsertNonZero InsertValues Update Delete instead
func Prepare[T sqlbuilder.Model](v *T) MutationDeprecated[T] {
	if _, ok := any(v).(sqlbuilder.Table); ok {
		panic(fmt.Errorf("prepare should be some data struct, but got %T", v))
	}

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

type MutationDeprecated[T sqlbuilder.Model] interface {
	IncludesZero(columns ...sqlbuilder.Column) Mutation[T]
	Values(valueSeq iter.Seq[*T], cols ...sqlbuilder.Column) Mutation[T]
	FromSelect(q Querier, cols ...sqlbuilder.Column) Mutation[T]

	ForDelete(opts ...OptionFunc) MutationMustWithWhere[T]
	ForUpdateSet(assignments ...sqlbuilder.Assignment) MutationMustWithWhere[T]

	MutationMustWithWhere[T]
}

type MutationMustWithWhere[T any] interface {
	Mutation[T]

	Where(where sqlfrag.Fragment) Mutation[T]
}

func InsertNonZero[T sqlbuilder.Model](value *T, zeroFieldsIncludes ...sqlbuilder.Column) Mutation[T] {
	if m, ok := any(value).(ModelWithCreationTime); ok {
		m.MarkCreatedAt()
	}

	return &mutation[T]{
		target: value,
		feature: feature{
			softDelete: true,
		},
		zeroFieldsIncludes: zeroFieldsIncludes,
	}
}

func UpdateNonZero[T sqlbuilder.Model](v *T, zeroFieldsIncludes ...sqlbuilder.Column) MutationMustWithWhere[T] {
	return &mutation[T]{
		target: v,
		feature: feature{
			softDelete: true,
		},
		zeroFieldsIncludes: zeroFieldsIncludes,
	}
}

func Insert[T sqlbuilder.Model](value *T, cols ...sqlbuilder.Column) Mutation[T] {
	if m, ok := any(value).(ModelWithCreationTime); ok {
		m.MarkCreatedAt()
	}

	return &mutation[T]{
		target: value,
		feature: feature{
			softDelete: true,
		},
		insertValues: &insertValues[T]{
			columns:  cols,
			valueSeq: slices.Values([]*T{value}),
		},
	}
}

func InsertValues[T sqlbuilder.Model](valueSeq iter.Seq[*T], cols ...sqlbuilder.Column) Mutation[T] {
	return &mutation[T]{
		target: new(T),
		feature: feature{
			softDelete: true,
		},
		insertValues: &insertValues[T]{
			columns:  cols,
			valueSeq: valueSeq,
		},
	}
}

func Update[T sqlbuilder.Model](v *T, cols ...sqlbuilder.Column) MutationMustWithWhere[T] {
	return &mutation[T]{
		target: v,
		feature: feature{
			softDelete: true,
		},
		update: &updateAction{
			columns: cols,
		},
	}
}

func Delete[T sqlbuilder.Model](opts ...OptionFunc) MutationMustWithWhere[T] {
	c := &mutation[T]{
		target:    new(T),
		forDelete: true,
		feature: feature{
			softDelete: true,
		},
	}

	for i := range opts {
		opts[i](c)
	}

	return c
}

type Mutation[T any] interface {
	WhereAnd(where sqlfrag.Fragment) Mutation[T]
	WhereOr(where sqlfrag.Fragment) Mutation[T]

	Apply(patchers ...MutationPatcher[T]) Mutation[T]

	OnConflict(cols sqlbuilder.ColumnSeq) Mutation[T]
	DoNothing() Mutation[T]
	DoUpdateSet(cols ...sqlbuilder.Column) Mutation[T]
	DoWith(func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition) Mutation[T]

	Returning(cols ...sqlfrag.Fragment) Mutation[T]
	Scan(recv any) Mutation[T]

	Save(ctx context.Context) error
}

type MutationPatcher[T any] interface {
	ApplyMutation(m Mutation[T]) Mutation[T]
}

type mutation[T any] struct {
	target               *T
	recv                 any
	zeroFieldsIncludes   []sqlbuilder.Column
	assignmentsForUpdate sqlbuilder.Assignments

	where sqlfrag.Fragment

	conflict              sqlbuilder.ColumnSeq
	onConflictDoWith      func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.Addition
	onConflictDoUpdateSet []sqlbuilder.Column

	fromSelect   *fromSelect
	update       *updateAction
	insertValues *insertValues[T]

	returning []sqlfrag.Fragment

	forDelete bool

	feature
}

func (c mutation[T]) Apply(patchers ...MutationPatcher[T]) Mutation[T] {
	var applied Mutation[T] = &c

	for _, p := range patchers {
		if p != nil {
			applied = p.ApplyMutation(applied)
		}
	}

	return applied
}

type updateAction struct {
	columns columns
}

type insertValues[T any] struct {
	columns  columns
	valueSeq iter.Seq[*T]
}

type columns []sqlbuilder.Column

func (cc columns) colsForMut(t sqlbuilder.Table) sqlbuilder.ColumnCollection {
	cols := sqlbuilder.Cols()

	if len(cc) > 0 {
		for _, c := range cc {
			if col := t.F(c.FieldName()); col != nil {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}
	} else {
		for col := range t.Cols() {
			if !sqlbuilder.GetColumnDef(col).AutoIncrement {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}
	}

	return cols
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

func (c mutation[T]) Values(valueSeq iter.Seq[*T], cols ...sqlbuilder.Column) Mutation[T] {
	c.insertValues = &insertValues[T]{
		columns:  cols,
		valueSeq: valueSeq,
	}
	return &c
}

func (c mutation[T]) FromSelect(q Querier, cols ...sqlbuilder.Column) Mutation[T] {
	c.fromSelect = &fromSelect{
		columns: cols,
		values:  q.AsSelect(),
	}
	return &c
}

func (c mutation[T]) ForDelete(fns ...OptionFunc) MutationMustWithWhere[T] {
	c.forDelete = true
	for i := range fns {
		fns[i](&c)
	}
	return &c
}

func (c mutation[T]) ForUpdateSet(assignments ...sqlbuilder.Assignment) MutationMustWithWhere[T] {
	c.assignmentsForUpdate = assignments
	return &c
}

func (c mutation[T]) Where(where sqlfrag.Fragment) Mutation[T] {
	c.where = where
	return &c
}

func (c mutation[T]) WhereAnd(where sqlfrag.Fragment) Mutation[T] {
	c.where = sqlbuilder.And(c.where, where)
	return &c
}

func (c mutation[T]) WhereOr(where sqlfrag.Fragment) Mutation[T] {
	c.where = sqlbuilder.Or(c.where, where)
	return &c
}

func (c mutation[T]) OnConflict(cols sqlbuilder.ColumnSeq) Mutation[T] {
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

func (c mutation[T]) Returning(cols ...sqlfrag.Fragment) Mutation[T] {
	if len(cols) != 0 {
		c.returning = cols
	} else {
		c.returning = make([]sqlfrag.Fragment, 0)
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
				t.F(f).Fragment("# = ?", notDeletedValue),
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

	var stmt sqlfrag.Fragment

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

func (c *mutation[T]) includeFieldNames() []string {
	zeroFieldsIncludes := make([]string, len(c.zeroFieldsIncludes))
	for i := range zeroFieldsIncludes {
		zeroFieldsIncludes[i] = c.zeroFieldsIncludes[i].FieldName()
	}
	return zeroFieldsIncludes
}

func (c *mutation[T]) insertOrUpdate(ctx context.Context, t sqlbuilder.Table, s Session) error {
	additions := make([]sqlbuilder.Addition, 0)

	if c.conflict != nil {
		onConflict := sqlbuilder.OnConflict(c.conflict)

		if onConflictDoWith := c.onConflictDoWith; onConflictDoWith != nil {
			additions = append(additions, onConflictDoWith(onConflict))
		} else {
			cols := c.onConflictDoUpdateSet
			if cols == nil {
				// FIXME ugly hack
				// sqlite will not RETURNING when ON CONFLICT DO NOTHING
				for col := range c.conflict.Cols() {
					cols = append(cols, col)
				}
			}

			assignments := make([]sqlbuilder.Assignment, len(cols))
			for idx, col := range cols {
				assignments[idx] = sqlbuilder.ColumnsAndValues(
					col, col.Fragment("EXCLUDED.?", sqlfrag.Const(col.Name())),
				)
			}

			onConflict = onConflict.DoUpdateSet(assignments...)
			additions = append(additions, onConflict)
		}
	}

	additions, hasReturning := c.withReturning(t, additions)

	var stmt sqlfrag.Fragment

	if where := c.buildWhere(t); where != nil {
		var assignmentsForUpdate []sqlbuilder.Assignment

		if c.update != nil {
			assignmentsForUpdate = toAssignmentsOf(c.target, c.update.columns.colsForMut(t))
		} else {
			assignmentsForUpdate = c.assignmentsForUpdate
			if len(assignmentsForUpdate) == 0 {
				assignmentsForUpdate = toAssignments(t, c.target, c.includeFieldNames()...)
			}
		}

		stmt = sqlbuilder.Update(t).Where(where, additions...).Set(assignmentsForUpdate...)
	} else if c.fromSelect != nil {
		stmt = sqlbuilder.Insert().Into(t, additions...).
			Values(
				sqlbuilder.PickColsByFieldNames(t, slicesx.Map(c.fromSelect.columns, func(e sqlbuilder.Column) string {
					return e.Name()
				})...),
				c.fromSelect.values,
			)
	} else if c.insertValues != nil {
		cols := c.insertValues.columns.colsForMut(t)

		stmt = sqlbuilder.Insert().Into(t, additions...).ValuesCollect(cols, func(yield func(any) bool) {
			for value := range c.insertValues.valueSeq {
				for sfv := range structs.AllFieldValue(context.Background(), value) {
					if col := cols.F(sfv.Field.FieldName); col != nil {
						v := sfv.Value.Interface()

						if m, ok := v.(ModelWithCreationTime); ok {
							m.MarkCreatedAt()
						}

						if !yield(v) {
							return
						}
					}
				}
			}
		})
	} else {
		cols, vals := toColumnsAndValues(t, c.target, c.includeFieldNames()...)
		stmt = sqlbuilder.Insert().Into(t, additions...).Values(cols, vals...)
	}

	return c.exec(ctx, s, stmt, hasReturning)
}

func (c *mutation[T]) exec(ctx context.Context, s Session, stmt sqlfrag.Fragment, hasReturning bool) error {
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
			additions = append(additions, sqlbuilder.Returning(sqlfrag.Const("*")))
		} else {
			additions = append(additions, sqlbuilder.Returning(sqlbuilder.MultiMayAutoAlias(c.returning...)))
		}
	}

	return additions, hasReturning
}

func toColumnsAndValues(t sqlbuilder.Table, m any, excludeFields ...string) (cols sqlbuilder.ColumnCollection, args []any) {
	cols = sqlbuilder.Cols()

	for sfv := range structs.AllFieldValueOmitZero(context.Background(), m, excludeFields...) {
		if col := t.F(sfv.Field.FieldName); col != nil {
			cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			args = append(args, sfv.Value.Interface())
		}
	}

	return
}

func toAssignments(t sqlbuilder.Table, m any, excludeFields ...string) (assignments sqlbuilder.Assignments) {
	for sfv := range structs.AllFieldValueOmitZero(context.Background(), m, excludeFields...) {
		if col := t.F(sfv.Field.FieldName); col != nil {
			assignments = append(assignments, sqlbuilder.CastColumn[any](col).By(sqlbuilder.Value(sfv.Value.Interface())))
		}
	}
	return
}

func toAssignmentsOf(m any, cc sqlbuilder.ColumnCollection) (assignments sqlbuilder.Assignments) {
	for sfv := range structs.AllFieldValue(context.Background(), m) {
		if col := cc.F(sfv.Field.FieldName); col != nil {
			assignments = append(assignments, sqlbuilder.CastColumn[any](col).By(sqlbuilder.Value(sfv.Value.Interface())))
		}
	}
	return
}
