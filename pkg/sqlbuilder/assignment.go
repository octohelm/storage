package sqlbuilder

import (
	"context"
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type Assignment interface {
	sqlfrag.Fragment

	SqlAssignment()
}

func ColumnsAndValues(columnOrColumns sqlfrag.Fragment, values ...any) Assignment {
	lenOfColumn := 1
	if canLen, ok := columnOrColumns.(interface{ Len() int }); ok {
		lenOfColumn = canLen.Len()
	}

	return &assignment{
		columnOrColumns: columnOrColumns,
		lenOfColumn:     lenOfColumn,
		values:          values,
	}
}

func ColumnsAndCollect(columnOrColumns sqlfrag.Fragment, seq iter.Seq[any]) Assignment {
	lenOfColumn := 1
	if canLen, ok := columnOrColumns.(interface{ Len() int }); ok {
		lenOfColumn = canLen.Len()
	}

	return &assignment{
		columnOrColumns: columnOrColumns,
		lenOfColumn:     lenOfColumn,
		valueSeq:        seq,
	}
}

type Assignments []Assignment

func (assignments Assignments) IsNil() bool {
	return len(assignments) == 0
}

func (assignments Assignments) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.Join(", ", sqlfrag.NonNil(slices.Values(assignments))).Frag(ctx)
}

type assignment struct {
	columnOrColumns sqlfrag.Fragment
	lenOfColumn     int
	values          []any
	valueSeq        iter.Seq[any]
}

func (assignment) SqlAssignment() {}

func (a *assignment) IsNil() bool {
	return a == nil || sqlfrag.IsNil(a.columnOrColumns) || (a.valueSeq == nil && len(a.values) == 0)
}

func (a *assignment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	useValues := TogglesFromContext(ctx).Is(ToggleUseValues)

	return func(yield func(string, []any) bool) {
		// (f_a,f_b)
		if useValues || a.valueSeq != nil {
			for q, args := range sqlfrag.Group(a.columnOrColumns).Frag(ContextWithToggles(ctx, Toggles{
				ToggleMultiTable: false,
			})) {
				if !yield(q, args) {
					return
				}
			}

			values := a.values

			if a.valueSeq != nil {
				values = slices.Collect(a.valueSeq)
			}

			// FROM
			if len(values) == 1 {
				if s, ok := values[0].(SelectStatement); ok {
					if !yield(" ", nil) {
						return
					}

					for q, args := range s.Frag(ctx) {
						if !yield(q, args) {
							return
						}
					}
					return
				}
			}

			if !yield(" VALUES ", nil) {
				return
			}

			valuesFragmentSeq := sqlfrag.Map(slices.Chunk(values, a.lenOfColumn), func(values []any) sqlfrag.Fragment {
				return sqlfrag.Pair("("+strings.Repeat(",?", len(values))[1:]+")", values...)
			})

			for q, args := range sqlfrag.Join(",", valuesFragmentSeq).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}

			return
		}

		for q, args := range a.columnOrColumns.Frag(ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: false,
		})) {
			if !yield(q, args) {
				return
			}
		}

		for q, args := range sqlfrag.Pair(" = ?", a.values[0]).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
