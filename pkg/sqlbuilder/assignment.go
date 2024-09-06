package sqlbuilder

import (
	"context"
	"iter"
	"math"
	"slices"
)

func WriteAssignments(e *Ex, assignments ...Assignment) {
	count := 0

	for i := range assignments {
		a := assignments[i]

		if IsNilExpr(a) {
			continue
		}

		if count > 0 {
			e.WriteQuery(", ")
		}

		e.WriteExpr(a)
		count++
	}
}

type Assignment interface {
	SqlExpr
	SqlAssignment()
}

func ColumnsAndValues(columnOrColumns SqlExpr, values ...any) Assignment {
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

func ColumnsAndCollect(columnOrColumns SqlExpr, seq iter.Seq[any]) Assignment {
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

type assignment struct {
	columnOrColumns SqlExpr
	lenOfColumn     int
	values          []any
	valueSeq        iter.Seq[any]
}

func (assignment) SqlAssignment() {}

func (a *assignment) IsNil() bool {
	return a == nil || IsNilExpr(a.columnOrColumns) || (a.valueSeq == nil && len(a.values) == 0)
}

func (a *assignment) Ex(ctx context.Context) *Ex {
	e := Expr("")

	useValues := TogglesFromContext(ctx).Is(ToggleUseValues)

	if useValues || a.valueSeq != nil {
		e.WriteGroup(func(e *Ex) {
			e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
				return a.columnOrColumns.Ex(ContextWithToggles(ctx, Toggles{
					ToggleMultiTable: false,
				}))
			}))
		})

		values := a.values

		if a.valueSeq != nil {
			values = slices.Collect(a.valueSeq)
		}

		if len(values) == 1 {
			if s, ok := values[0].(SelectStatement); ok {
				e.WriteQueryByte(' ')
				e.WriteExpr(s)
				return e.Ex(ctx)
			}
		}

		e.WriteQuery(" VALUES ")

		groupCount := int(math.Round(float64(len(values)) / float64(a.lenOfColumn)))

		for i := 0; i < groupCount; i++ {
			if i > 0 {
				e.WriteQueryByte(',')
			}

			e.WriteGroup(func(e *Ex) {
				for j := 0; j < a.lenOfColumn; j++ {
					e.WriteHolder(j)
				}
			})
		}

		e.AppendArgs(values...)

		return e.Ex(ctx)
	}

	e.WriteExpr(ExprBy(func(ctx context.Context) *Ex {
		return a.columnOrColumns.Ex(ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: false,
		}))
	}))

	e.WriteQuery(" = ?")
	e.AppendArgs(a.values[0])

	return e.Ex(ctx)
}
