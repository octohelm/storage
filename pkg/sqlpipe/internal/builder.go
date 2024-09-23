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
	"github.com/octohelm/storage/pkg/sqltype"
	sqltypetime "github.com/octohelm/storage/pkg/sqltype/time"
	"github.com/octohelm/x/reflect"
)

func BuildStmt[M sqlbuilder.Model](ctx context.Context, patchers ...StmtPatcher[M]) sqlfrag.Fragment {
	b := &Builder[M]{}
	if flags, ok := FlagsContext.MayFrom(ctx); ok {
		b.Flags = flags
	}

	return b.ApplyPatchers(ctx, patchers...).(StmtCreator).BuildStmt(ctx)
}

func ApplyStmt[M sqlbuilder.Model](ctx context.Context, b StmtBuilder[M], patchers ...StmtPatcher[M]) StmtBuilder[M] {
	return b.(*Builder[M]).ApplyPatchers(ctx, patchers...)
}

func CollectStmt[M sqlbuilder.Model](ctx context.Context, patchers ...StmtPatcher[M]) iter.Seq2[string, []any] {
	return BuildStmt(ctx, patchers...).Frag(ctx)
}

type StmtBuilder[M sqlbuilder.Model] interface {
	WithSource(source sqlfrag.Fragment) StmtBuilder[M]

	WithDefaultProjects(projects ...sqlfrag.Fragment) StmtBuilder[M]
	WithProjects(projects ...sqlfrag.Fragment) StmtBuilder[M]

	WithAdditions(additions ...sqlbuilder.Addition) StmtBuilder[M]
	WithoutAddition(omit func(a sqlbuilder.Addition) bool) StmtBuilder[M]

	T(ctx context.Context, x any) sqlbuilder.Table

	WithFlags(flags Flags) StmtBuilder[M]
}

type StmtCreator interface {
	BuildStmt(ctx context.Context) sqlfrag.Fragment
}

type Builder[M sqlbuilder.Model] struct {
	Source          sqlfrag.Fragment
	Additions       []sqlbuilder.Addition
	Projects        []sqlfrag.Fragment
	DefaultProjects []sqlfrag.Fragment
	Flags
}

func (s Builder[M]) WithFlags(flags Flags) StmtBuilder[M] {
	s.Flags = s.Flags.Merge(flags)
	return &s
}

func (s Builder[M]) WithSource(table sqlfrag.Fragment) StmtBuilder[M] {
	s.Source = table
	return &s
}

func (s Builder[M]) WithAdditions(additions ...sqlbuilder.Addition) StmtBuilder[M] {
	s.Additions = append(s.Additions, additions...)
	return &s
}

func (s Builder[M]) WithoutAddition(omit func(a sqlbuilder.Addition) bool) StmtBuilder[M] {
	additions := make([]sqlbuilder.Addition, 0, len(s.Additions))

	for _, a := range s.Additions {
		if !omit(a) {
			additions = append(additions, a)
		}
	}

	s.Additions = additions

	return &s
}

func (s Builder[M]) WithProjects(projects ...sqlfrag.Fragment) StmtBuilder[M] {
	s.Projects = projects
	return &s
}

func (s Builder[M]) WithDefaultProjects(projects ...sqlfrag.Fragment) StmtBuilder[M] {
	if len(s.DefaultProjects) == 0 {
		s.DefaultProjects = projects
	}
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
	if s.IncludesAll() {
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
			if x, ok := any(value).(sqltype.WithCreationTime); ok {
				x.MarkCreatedAt()
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

	var where sqlfrag.Fragment

	for _, a := range s.Additions {
		if a.AdditionType() == sqlbuilder.AdditionWhere {
			where = a
			continue
		}
		additions = append(additions, a)
	}

	if where != nil {
		if w := s.PatchWhere(ctx, where); !sqlfrag.IsNil(w) {
			additions = append(additions, sqlbuilder.Where(w))
		}
	} else {
		if w := s.PatchWhere(ctx, nil); !sqlfrag.IsNil(w) {
			additions = append(additions, sqlbuilder.Where(w))
		} else {
			if s.WhereRequired() {
				return sqlfrag.Empty()
			}
		}
	}

	var project sqlfrag.Fragment
	if projects := s.prepareProjects(); len(projects) > 0 {
		project = sqlfrag.JoinValues(", ", projects...)
	}

	return sqlbuilder.Select(project).From(from, additions...)
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

func (s *Builder[M]) ApplyPatchers(ctx context.Context, patchers ...StmtPatcher[M]) StmtBuilder[M] {
	var b StmtBuilder[M] = s
	for _, patcher := range patchers {
		b = patcher.ApplyStmt(ctx, b)
	}
	return b
}
