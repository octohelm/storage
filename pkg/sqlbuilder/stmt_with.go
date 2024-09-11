package sqlbuilder

import (
	"context"
	"iter"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type BuildSubQuery func(table Table) sqlfrag.Fragment

func WithRecursive(t Table, build BuildSubQuery) *WithStmt {
	return With(t, build, "RECURSIVE")
}

func With(t Table, build BuildSubQuery, modifiers ...string) *WithStmt {
	return (&WithStmt{modifiers: modifiers}).With(t, build)
}

type WithStmt struct {
	modifiers []string
	tables    []Table
	asList    []BuildSubQuery
	statement func(tables ...Table) sqlfrag.Fragment
}

func (w *WithStmt) IsNil() bool {
	return w == nil || len(w.tables) == 0 || len(w.asList) == 0 || w.statement == nil
}

func (w WithStmt) With(t Table, build BuildSubQuery) *WithStmt {
	w.tables = append(w.tables, t)
	w.asList = append(w.asList, build)
	return &w
}

func (w WithStmt) Exec(statement func(tables ...Table) sqlfrag.Fragment) *WithStmt {
	w.statement = statement
	return &w
}

func (w *WithStmt) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("WITH", nil) {
			return
		}

		if len(w.modifiers) > 0 {
			if !yield(" "+strings.Join(w.modifiers, " "), nil) {
				return
			}
		}

		for i, t := range w.tables {
			if i > 0 {
				if !yield(",", nil) {
					return
				}
			}

			if !yield("\n", nil) {
				return
			}

			for q, args := range t.Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
			for q, args := range sqlfrag.Group(sqlfrag.Join(",", sqlfrag.Map(t.Cols(), func(col Column) sqlfrag.Fragment {
				return col
			}))).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}

			if !yield("\n  AS ", nil) {
				return
			}

			build := w.asList[i]

			for q, args := range sqlfrag.Group(build(t)).Frag(ctx) {
				if !yield(q, args) {
					return
				}
			}
		}

		if !yield("\n", nil) {
			return
		}

		for q, args := range w.statement(w.tables...).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
