package internal

import (
	"context"
	"errors"
	"iter"
	"slices"

	reflectx "github.com/octohelm/x/reflect"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqltype"
)

type DeleteType uint

const (
	DeleteTypeNone DeleteType = iota
	DeleteTypeHard
	DeleteTypeSoft
)

type Mutation[M sqlbuilder.Model] struct {
	ForDelete DeleteType
	ForUpdate bool

	From any

	OmitZero OmitZero[M]
	Strict   Strict[M]

	Assignments []sqlbuilder.Assignment
	Values      iter.Seq[*M]
}

type OmitZero[M sqlbuilder.Model] struct {
	Enabled bool
	Exclude []modelscoped.Column[M]
}

type Strict[M sqlbuilder.Model] struct {
	Omit    bool
	Columns []modelscoped.Column[M]
}

func (s *Strict[M]) StrictColumnCollection(t sqlbuilder.Table) sqlbuilder.ColumnCollection {
	cols := sqlbuilder.Cols()

	if len(s.Columns) > 0 {
		if s.Omit {
			excludes := sqlbuilder.Cols()
			for _, c := range s.Columns {
				if col := t.F(c.FieldName()); col != nil {
					excludes.(sqlbuilder.ColumnCollectionManger).AddCol(col)
				}
			}

			// all
			for col := range t.Cols() {
				def := sqlbuilder.GetColumnDef(col)

				if def.DeprecatedActions != nil || def.AutoIncrement {
					continue
				}

				if c := excludes.F(col.FieldName()); c == nil {
					cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
				}
			}

			return cols
		}

		for _, c := range s.Columns {
			if col := t.F(c.FieldName()); col != nil {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}

		return cols
	}

	for col := range t.Cols() {
		def := sqlbuilder.GetColumnDef(col)

		if def.DeprecatedActions != nil || def.AutoIncrement {
			continue
		}

		cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
	}

	return cols
}

func (m *Mutation[M]) IsNil() bool {
	return false
}

func (m Mutation[M]) WithAssignments(assignments ...sqlbuilder.Assignment) *Mutation[M] {
	m.Assignments = append(m.Assignments, assignments...)
	return &m
}

func (m *Mutation[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		return
	}
}

func (m *Mutation[M]) PrepareColumnCollectionForInsert(t sqlbuilder.Table) sqlbuilder.ColumnCollection {
	return m.Strict.StrictColumnCollection(t)
}

func (m *Mutation[M]) PrepareAssignments(ctx context.Context, t sqlbuilder.Table) iter.Seq[sqlbuilder.Assignment] {
	if m.Assignments != nil {
		return slices.Values(m.Assignments)
	}

	values := slices.Collect(func(yield func(*M) bool) {
		for value := range m.Values {
			if x, ok := any(value).(sqltype.WithModificationTime); ok {
				x.MarkModifiedAt()
			}

			if !yield(value) {
				return
			}
		}
	})
	if len(values) == 0 {
		panic(errors.New("assigment required at least one value"))
	}
	if len(values) > 1 {
		panic(errors.New("assigment only support single value"))
	}

	if m.OmitZero.Enabled {
		includes := sqlbuilder.Cols()

		for _, c := range m.OmitZero.Exclude {
			if col := t.F(c.FieldName()); col != nil {
				includes.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}

		return func(yield func(sqlbuilder.Assignment) bool) {
			for sfv := range structs.AllFieldValue(ctx, values[0]) {
				if includes.F(sfv.Field.FieldName) != nil || !reflectx.IsEmptyValue(sfv.Value) {
					if col := t.F(sfv.Field.FieldName); col != nil {
						if !(yield(sqlbuilder.CastColumn[any](col).By(sqlbuilder.Value(sfv.Value.Interface())))) {
							return
						}
					}
				}
			}
		}
	}

	if m.Strict.Columns != nil {
		cols := m.Strict.StrictColumnCollection(t)

		return func(yield func(sqlbuilder.Assignment) bool) {
			for sfv := range structs.AllFieldValue(ctx, values[0]) {
				if col := cols.F(sfv.Field.FieldName); col != nil {
					if !(yield(sqlbuilder.CastColumn[any](col).By(sqlbuilder.Value(sfv.Value.Interface())))) {
						return
					}
				}
			}
		}
	}

	return nil
}
