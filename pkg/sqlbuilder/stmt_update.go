package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/pkg/errors"
)

var UpdateNeedLimitByWhere = errors.New("no where limit for update")

func Update(table Table, modifiers ...string) *StmtUpdate {
	return &StmtUpdate{
		table:     table,
		modifiers: modifiers,
	}
}

type StmtUpdate struct {
	table     Table
	from      Table
	modifiers []string

	assignments Assignments
	additions   Additions
}

func (s *StmtUpdate) IsNil() bool {
	return s == nil || sqlfrag.IsNil(s.table) || len(s.assignments) == 0
}

func (s StmtUpdate) Set(assignments ...Assignment) *StmtUpdate {
	s.assignments = assignments
	return &s
}

func (s StmtUpdate) From(table Table, additions ...Addition) *StmtUpdate {
	s.from = table

	if len(additions) > 0 {
		s.additions = append(s.additions, additions...)
	}
	return &s
}

func (s StmtUpdate) Where(c sqlfrag.Fragment, additions ...Addition) *StmtUpdate {
	s.additions = []Addition{Where(c)}
	if len(additions) > 0 {
		s.additions = append(s.additions, additions...)
	}
	return &s
}

func (s *StmtUpdate) Frag(ctx context.Context) iter.Seq2[string, []any] {
	if s.from != nil {
		ctx = ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: true,
		})
	}

	return func(yield func(string, []any) bool) {
		if !yield("UPDATE", nil) {
			return
		}

		for i := range s.modifiers {
			if !yield(" "+s.modifiers[i], nil) {
				return
			}
		}

		if !yield(" ", nil) {
			return
		}

		for q, args := range s.table.Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}

		if s.assignments != nil {
			if !yield(" SET ", nil) {
				return
			}

			for q, args := range s.assignments.Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
		}

		if s.from != nil {
			if !yield(" FROM ", nil) {
				return
			}
			for q, args := range s.from.Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
		}

		for q, args := range s.additions.Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
