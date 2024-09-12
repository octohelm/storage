package sqlbuilder

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type OnConflictAddition interface {
	Addition

	DoNothing() OnConflictAddition
	DoUpdateSet(assignments ...Assignment) OnConflictAddition
}

func OnConflict(columns ColumnSeq) OnConflictAddition {
	return &onConflict{
		columns: columns,
	}
}

type onConflict struct {
	columns     ColumnSeq
	doNothing   bool
	assignments []Assignment
}

func (onConflict) AdditionType() AdditionType {
	return AdditionOnConflict
}

func (o onConflict) DoNothing() OnConflictAddition {
	o.doNothing = true
	return &o
}

func (o onConflict) DoUpdateSet(assignments ...Assignment) OnConflictAddition {
	o.assignments = assignments
	return &o
}

func (o *onConflict) IsNil() bool {
	return o == nil || o.columns == nil || (!o.doNothing && len(o.assignments) == 0)
}

func (o *onConflict) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield("ON CONFLICT ", nil) {
			return
		}

		columnSeq := sqlfrag.Join(",", sqlfrag.Map(o.columns.Cols(), func(col Column) sqlfrag.Fragment {
			return col
		}))

		for q, args := range sqlfrag.InlineBlock(columnSeq).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}

		if !yield(" DO ", nil) {
			return
		}

		if o.doNothing {
			if !yield("NOTHING", nil) {
				return
			}
			return
		}

		if !yield("UPDATE SET ", nil) {
			return
		}

		for q, args := range sqlfrag.Join(", ", sqlfrag.NonNil(slices.Values(o.assignments))).Frag(ctx) {
			if !yield(q, args) {
				return
			}
		}
	}
}
