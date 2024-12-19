package internal

import (
	"context"
	"fmt"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func ToString(s sqlfrag.Fragment) string {
	q, args := sqlfrag.Collect(context.Background(), s)
	return fmt.Sprintf("%s | %v", q, args)
}

func ColumnsByStruct(v any) sqlfrag.Fragment {
	return sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
		return func(yield func(string, []any) bool) {
			i := 0

			for fieldValue := range structs.AllFieldValue(ctx, v) {
				if i > 0 {
					if !yield(", ", nil) {
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

func fixAdditions(additions []sqlbuilder.Addition) []sqlbuilder.Addition {
	hasAdditionReturning := false

	finalAdditions := make([]sqlbuilder.Addition, 0, len(additions))

	for _, a := range slices.SortedFunc(slices.Values(additions), sqlbuilder.CompareAddition) {
		switch a.AdditionType() {
		case sqlbuilder.AdditionReturning:
			hasAdditionReturning = true
		case sqlbuilder.AdditionLimit:
			// drop limit when returning exists
			if hasAdditionReturning {
				continue
			}
		default:

		}

		finalAdditions = append(finalAdditions, a)
	}

	return finalAdditions
}
