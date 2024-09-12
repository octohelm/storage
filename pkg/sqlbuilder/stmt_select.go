package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type SelectStatement interface {
	sqlfrag.Fragment

	selectStatement()
}

func Select(sqlExpr sqlfrag.Fragment, modifiers ...sqlfrag.Fragment) *StmtSelect {
	return &StmtSelect{
		projects:  sqlExpr,
		modifiers: modifiers,
	}
}

type StmtSelect struct {
	SelectStatement

	table     sqlfrag.Fragment
	modifiers []sqlfrag.Fragment
	projects  sqlfrag.Fragment
	additions Additions
}

func (s *StmtSelect) IsNil() bool {
	return s == nil
}

func (s StmtSelect) From(table sqlfrag.Fragment, additions ...Addition) *StmtSelect {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtSelect) Frag(ctx context.Context) iter.Seq2[string, []any] {
	multiTable := false

	for i := range s.additions {
		addition := s.additions[i]
		if sqlfrag.IsNil(addition) {
			continue
		}

		if addition.AdditionType() == AdditionJoin {
			multiTable = true
		}
	}

	if multiTable {
		ctx = ContextWithToggles(ctx, Toggles{
			ToggleMultiTable: multiTable,
		})
	}

	return func(yield func(string, []any) bool) {
		if !yield("\nSELECT", nil) {
			return
		}

		for _, m := range s.modifiers {
			for q, args := range m.Frag(ctx) {
				if !yield(" "+q, args) {
					return
				}
			}
		}

		if !yield(" ", nil) {
			return
		}

		projects := s.projects

		if sqlfrag.IsNil(s.projects) {
			projects = sqlfrag.Const("*")
		}

		for q, args := range projects.Frag(ContextWithToggles(ctx, Toggles{
			ToggleInProject: true,
		})) {
			if !yield(q, args) {
				return
			}
		}

		if !sqlfrag.IsNil(s.table) {
			if !yield("\nFROM ", nil) {
				return
			}

			for q, args := range s.table.Frag(ctx) {
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

func ForUpdate() *OtherAddition {
	return AsAddition(sqlfrag.Pair("FOR UPDATE"))
}
