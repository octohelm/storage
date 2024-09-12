package sqlbuilder

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Delete() *StmtDelete {
	return &StmtDelete{}
}

type StmtDelete struct {
	table     Table
	additions Additions
}

func (s *StmtDelete) IsNil() bool {
	return s == nil || sqlfrag.IsNil(s.table)
}

func (s StmtDelete) From(table Table, additions ...Addition) *StmtDelete {
	s.table = table
	s.additions = additions
	return &s
}

func (s *StmtDelete) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("\nDELETE FROM ", nil) {
			return
		}

		for q, args := range s.table.Frag(ctx) {
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
