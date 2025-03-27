package internal

import (
	"context"
	"database/sql/driver"
	"iter"
	"time"

	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
	"github.com/octohelm/storage/pkg/sqltype"
	sqltypetime "github.com/octohelm/storage/pkg/sqltype/time"
	"github.com/octohelm/x/reflect"
	slicesx "github.com/octohelm/x/slices"
)

func BuildStmt[M sqlbuilder.Model](ctx context.Context, patchers ...StmtPatcher[M]) sqlfrag.Fragment {
	b := &Builder[M]{}
	if f, ok := FlagContext.MayFrom(ctx); ok {
		b.Flag = f
	}
	return b.ApplyPatchers(ctx, patchers...).BuildStmt(ctx)
}

func ApplyStmt[M sqlbuilder.Model](ctx context.Context, b *Builder[M], patchers ...StmtPatcher[M]) *Builder[M] {
	return b.ApplyPatchers(ctx, patchers...)
}

func CollectStmt[M sqlbuilder.Model](ctx context.Context, patchers ...StmtPatcher[M]) iter.Seq2[string, []any] {
	return BuildStmt(ctx, patchers...).Frag(ctx)
}

type StmtCreator interface {
	BuildStmt(ctx context.Context) sqlfrag.Fragment
}

type Builder[M sqlbuilder.Model] struct {
	flags.Flag

	Source     sqlfrag.Fragment
	TableJoins []sqlbuilder.JoinAddition

	Orders []sqlbuilder.Order

	Projects        []sqlfrag.Fragment
	DefaultProjects []sqlfrag.Fragment

	DistinctOn []sqlfrag.Fragment

	Pager sqlbuilder.Addition

	Additions []sqlbuilder.Addition
}

func (s Builder[M]) WithFlag(f flags.Flag) *Builder[M] {
	s.Flag = s.Flag | f
	return &s
}

func (s Builder[M]) WithSource(table sqlfrag.Fragment) *Builder[M] {
	s.Source = table
	return &s
}

func (s Builder[M]) WithTableJoins(tableJoins ...sqlbuilder.JoinAddition) *Builder[M] {
	s.TableJoins = append(tableJoins, s.TableJoins...)
	return &s
}

func (s Builder[M]) WithOrders(orders ...sqlbuilder.Order) *Builder[M] {
	s.Orders = append(orders, s.Orders...)
	return &s
}

func (s Builder[M]) WithAdditions(additions ...sqlbuilder.Addition) *Builder[M] {
	s.Additions = append(s.Additions, additions...)
	return &s
}

func (s Builder[M]) WithDistinctOn(on ...sqlfrag.Fragment) *Builder[M] {
	s.DistinctOn = on
	return &s
}

func (s Builder[M]) WithProjects(projects ...sqlfrag.Fragment) *Builder[M] {
	s.Projects = projects
	return &s
}

func (s Builder[M]) WithDefaultProjects(projects ...sqlfrag.Fragment) *Builder[M] {
	if len(s.DefaultProjects) == 0 {
		s.DefaultProjects = projects
	}
	return &s
}

func (s Builder[M]) WithPager(pager sqlbuilder.Addition) *Builder[M] {
	s.Pager = pager
	return &s
}

func (s *Builder[M]) BuildStmt(ctx context.Context) sqlfrag.Fragment {
	switch x := s.Source.(type) {
	case *Mutation[M]:
		if x.ForDelete != DeleteTypeNone {
			return s.buildDelete(ctx, x)
		}
		if x.ForUpdate {
			return s.buildUpdate(ctx, x)
		}
		return s.buildInsert(ctx, x)
	default:
		return s.buildSelect(ctx)
	}
}

func (s *Builder[M]) prepareProjects() []sqlfrag.Fragment {
	if len(s.Projects) == 0 {
		return s.DefaultProjects
	}
	return s.Projects
}

func (s *Builder[M]) PatchWhere(ctx context.Context, where sqlfrag.Fragment) sqlfrag.Fragment {
	if s.Is(flags.IncludesAll) {
		return where
	}

	m := new(M)

	if soft, ok := any(m).(sqltype.WithSoftDelete); ok {
		t := s.T(ctx, m)
		f, notDeletedValue := soft.SoftDeleteFieldAndZeroValue()

		return sqlbuilder.And(
			where,
			t.F(f).Fragment("# = ?", notDeletedValue),
		)
	}

	return where
}

func (s *Builder[M]) buildDelete(ctx context.Context, mut *Mutation[M]) sqlfrag.Fragment {
	m := new(M)

	t := s.T(ctx, m)

	additions := s.Additions

	if projects := s.prepareProjects(); len(projects) > 0 {
		additions = append(additions, sqlbuilder.Returning(sqlfrag.JoinValues(", ", projects...)))
	}

	if mut.ForDelete == DeleteTypeSoft {
		if soft, ok := any(m).(sqltype.WithSoftDelete); ok {
			if x, ok := any(m).(sqltype.DeletedAtMarker); ok {
				x.MarkDeletedAt()
			}

			f, _ := soft.SoftDeleteFieldAndZeroValue()

			var softDeleteValue driver.Value
			if v, ok := ctx.(sqltype.SoftDeleteValueGetter); ok {
				softDeleteValue = v.GetDeletedAt()
			} else {
				softDeleteValue = sqltypetime.Timestamp(time.Now())
			}

			col := t.F(f)

			return sqlbuilder.Update(t).Where(nil, fixAdditions(additions)...).Set(
				sqlbuilder.ColumnsAndValues(col, softDeleteValue),
			)
		}
	}

	return sqlbuilder.Delete().From(t, fixAdditions(additions)...)
}

func (s *Builder[M]) buildUpdate(ctx context.Context, mut *Mutation[M]) *sqlbuilder.StmtUpdate {
	t := s.T(ctx, new(M))

	additions := s.Additions

	if projects := s.prepareProjects(); len(projects) > 0 {
		additions = append(additions, sqlbuilder.Returning(sqlfrag.JoinValues(", ", projects...)))
	}

	update := sqlbuilder.Update(t).Where(nil, fixAdditions(additions)...)

	if mut.From != nil {
		update = update.From(s.T(ctx, mut.From))
	}

	return update.SetBy(mut.PrepareAssignments(ctx, t))
}

func (s *Builder[M]) buildInsert(ctx context.Context, m *Mutation[M]) sqlfrag.Fragment {
	t := s.T(ctx, new(M))

	additions := s.Additions

	if projects := s.prepareProjects(); len(projects) > 0 {
		additions = append(additions, sqlbuilder.Returning(sqlfrag.JoinValues(", ", projects...)))
	}

	if m.OmitZero {
		includes := sqlbuilder.Cols()

		for _, c := range m.OmitZeroExclude {
			if col := t.F(c.FieldName()); col != nil {
				includes.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}

		orderedCols := sqlbuilder.Cols()
		values := make([]any, 0)

		for value := range m.Values {
			if x, ok := any(value).(sqltype.WithCreationTime); ok {
				x.MarkCreatedAt()
			}

			for sfv := range structs.AllFieldValue(ctx, value) {
				if includes.F(sfv.Field.FieldName) != nil || !reflect.IsEmptyValue(sfv.Value) {
					if col := t.F(sfv.Field.FieldName); col != nil {
						orderedCols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
						values = append(values, sfv.Value.Interface())
					}
				}
			}
			break
		}

		if len(values) == 0 {
			return sqlfrag.Empty()
		}

		return sqlbuilder.Insert().Into(t, fixAdditions(additions)...).Values(orderedCols, values...)
	}

	cols := m.PrepareColumnCollectionForInsert(t)

	if m.Values == nil {
		if m.From == nil {
			return nil
		}

		return sqlbuilder.Insert().Into(t, fixAdditions(additions)...).Values(cols, m.From)
	}

	orderedCols := sqlbuilder.Cols()

	for value := range m.Values {
		for sfv := range structs.AllFieldValue(ctx, value) {
			if col := cols.F(sfv.Field.FieldName); col != nil {
				orderedCols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}
		break
	}

	if orderedCols.Len() == 0 {
		return sqlfrag.Empty()
	}

	return sqlbuilder.Insert().Into(t, fixAdditions(additions)...).ValuesCollect(orderedCols, func(yield func(any) bool) {
		for value := range m.Values {
			if canSetModification, ok := any(value).(sqltype.WithModificationTime); ok {
				canSetModification.MarkModifiedAt()
			} else if canSetCreationTime, ok := any(value).(sqltype.WithCreationTime); ok {
				canSetCreationTime.MarkCreatedAt()
			}

			for sfv := range structs.AllFieldValue(ctx, value) {
				if col := cols.F(sfv.Field.FieldName); col != nil {
					fv := sfv.Value.Interface()
					if !yield(fv) {
						return
					}
				}
			}
		}
	})
}

func (s *Builder[M]) buildSelect(ctx context.Context) sqlfrag.Fragment {
	from := s.Source

	if s.Source == nil {
		from = s.T(ctx, new(M))
	}

	additions := make([]sqlbuilder.Addition, 0, len(s.Additions))

	if len(s.TableJoins) > 0 {
		additions = append(additions, slicesx.Map(s.TableJoins, func(a sqlbuilder.JoinAddition) sqlbuilder.Addition {
			return a
		})...)
	}

	var where sqlfrag.Fragment

	for _, a := range s.Additions {
		switch a.AdditionType() {
		case sqlbuilder.AdditionWhere:
			where = a
			continue
		default:
		}

		additions = append(additions, a)

	}

	// patch soft_delete field filter if need
	if w := s.PatchWhere(ctx, where); !sqlfrag.IsNil(w) {
		where = w
		additions = append(additions, sqlbuilder.Where(w))
	}

	if s.Is(flags.WhereOrPagerRequired) {
		if sqlfrag.IsNil(where) && s.Pager == nil {
			return sqlfrag.Empty()
		}
	}

	orders := s.Orders

	var project sqlfrag.Fragment
	if projects := s.prepareProjects(); len(projects) > 0 {
		project = sqlfrag.JoinValues(", ", projects...)
	}

	var modifiers []sqlfrag.Fragment
	if s.DistinctOn != nil {
		modifiers = append(modifiers, sqlbuilder.DistinctOn(s.DistinctOn...))

		orders = append(slicesx.Map(s.DistinctOn, func(col sqlfrag.Fragment) sqlbuilder.Order {
			return sqlbuilder.DefaultOrder(col)
		}), orders...)
	}

	if len(orders) > 0 {
		if !s.Is(flags.WithoutSorter) {
			additions = append(additions, sqlbuilder.OrderBy(orders...))
		}
	}

	if s.Pager != nil {
		if !s.Is(flags.WithoutPager) {
			additions = append(additions, s.Pager)
		}
	}

	return sqlbuilder.Select(project, modifiers...).From(from, additions...)
}

func (s *Builder[M]) T(ctx context.Context, m any) sqlbuilder.Table {
	if m == nil {
		m = new(M)
	}

	sess := session.For(ctx, m)
	if sess != nil {
		return sess.T(m)
	}
	return sqlbuilder.TableFromModel(m)
}

func (s *Builder[M]) ApplyPatchers(ctx context.Context, patchers ...StmtPatcher[M]) *Builder[M] {
	var b *Builder[M] = s
	for _, patcher := range patchers {
		b = patcher.ApplyStmt(ctx, b)
	}
	return b
}
