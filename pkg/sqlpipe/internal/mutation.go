package internal

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	reflectx "github.com/octohelm/x/reflect"
	"github.com/pkg/errors"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
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
	OmitZero  bool

	From            any
	StrictColumns   []modelscoped.Column[M]
	OmitZeroExclude []modelscoped.Column[M]
	Assignments     []sqlbuilder.Assignment
	Values          iter.Seq[*M]
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
	cols := sqlbuilder.Cols()

	if len(m.StrictColumns) > 0 {
		for _, c := range m.StrictColumns {
			if col := t.F(c.FieldName()); col != nil {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}
	} else {
		for col := range t.Cols() {
			def := sqlbuilder.GetColumnDef(col)

			if def.DeprecatedActions != nil {
				continue
			}

			if !def.AutoIncrement {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}
	}

	return cols
}

func (m *Mutation[M]) PrepareAssignments(ctx context.Context, t sqlbuilder.Table) iter.Seq[sqlbuilder.Assignment] {
	if m.Assignments != nil {
		return slices.Values(m.Assignments)
	}

	values := slices.Collect(m.Values)
	if len(values) == 0 {
		panic(errors.New("assigment required at least one value"))
	}
	if len(values) > 1 {
		panic(errors.New("assigment only support single value"))
	}

	if m.OmitZero {
		includes := sqlbuilder.Cols()

		for _, c := range m.OmitZeroExclude {
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

	if m.StrictColumns != nil {
		cols := sqlbuilder.Cols()

		for _, c := range m.StrictColumns {
			if col := t.F(c.FieldName()); col != nil {
				cols.(sqlbuilder.ColumnCollectionManger).AddCol(col)
			}
		}

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
