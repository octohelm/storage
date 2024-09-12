package sqlbuilder

import (
	"context"
	"iter"
)

func Insert(modifiers ...string) *StmtInsert {
	return &StmtInsert{
		modifiers: modifiers,
	}
}

type StmtInsert struct {
	table       Table
	modifiers   []string
	assignments Assignments
	additions   Additions
}

func (s StmtInsert) Into(table Table, additions ...Addition) *StmtInsert {
	s.table = table
	s.additions = additions
	return &s
}

func (s StmtInsert) Values(cols ColumnCollection, values ...any) *StmtInsert {
	s.assignments = Assignments{ColumnsAndValues(cols, values...)}
	return &s
}

func (s StmtInsert) ValuesCollect(cols ColumnCollection, seq iter.Seq[any]) *StmtInsert {
	s.assignments = Assignments{ColumnsAndCollect(cols, seq)}
	return &s
}

func (s *StmtInsert) IsNil() bool {
	return s == nil || s.table == nil || len(s.assignments) == 0
}

func (s *StmtInsert) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("\nINSERT", nil) {
			return
		}

		for i := range s.modifiers {
			if !yield(" "+s.modifiers[i], nil) {
				return
			}
		}

		if !yield(" INTO ", nil) {
			return
		}

		for q, args := range s.table.Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}

		if !yield(" ", nil) {
			return
		}

		for q, args := range s.assignments.Frag(ContextWithToggles(ctx, Toggles{
			ToggleUseValues: true,
		})) {
			if !yield(q, args) {
				return
			}
		}

		for q, args := range s.additions.Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
