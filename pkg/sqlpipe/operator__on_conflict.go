package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

func OnConflictDoNothing[M sqlbuilder.Model](cols modelscoped.ColumnSeq[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorOnConflict, func(src Source[M]) Source[M] {
		return &onConflictSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			cols: cols,
		}
	})
}

func OnConflictDoUpdateSet[M sqlbuilder.Model](cols modelscoped.ColumnSeq[M], toUpdates ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorOnConflict, func(src Source[M]) Source[M] {
		return &onConflictSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			cols:    cols,
			updates: toUpdates,
		}
	})
}

func OnConflictDoWith[M sqlbuilder.Model](
	cols modelscoped.ColumnSeq[M],
	with func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.OnConflictAddition,
) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorOnConflict, func(src Source[M]) Source[M] {
		return &onConflictSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			cols: cols,
			with: with,
		}
	})
}

type onConflictSource[M sqlbuilder.Model] struct {
	Embed[M]

	cols    modelscoped.ColumnSeq[M]
	updates []modelscoped.Column[M]
	with    func(onConflictAddition sqlbuilder.OnConflictAddition) sqlbuilder.OnConflictAddition
}

func (s *onConflictSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *onConflictSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return s.Underlying.ApplyStmt(
		ctx,
		b.WithAdditions(s.toOnConflictAddition(ctx, s.GetFlag(ctx))),
	)
}

func (s *onConflictSource[M]) toOnConflictAddition(ctx context.Context, f flags.Flag) sqlbuilder.OnConflictAddition {
	if s.with != nil {
		return s.with(sqlbuilder.OnConflict(s.cols))
	}

	if len(s.updates) > 0 {
		assignments := make([]sqlbuilder.Assignment, len(s.updates))

		for idx, col := range s.updates {
			assignments[idx] = sqlbuilder.ColumnsAndValues(
				col, col.Fragment("EXCLUDED.?", sqlfrag.Const(col.Name())),
			)
		}

		return sqlbuilder.OnConflict(s.cols).DoUpdateSet(assignments...)
	}

	if f.Is(flags.ForReturning) {
		for col := range s.cols.Cols() {
			return sqlbuilder.OnConflict(s.cols).DoUpdateSet(sqlbuilder.ColumnsAndValues(
				col, col.Fragment("EXCLUDED.?", sqlfrag.Const(col.Name())),
			))
		}
	}

	return sqlbuilder.OnConflict(s.cols).DoNothing()
}

func (s *onConflictSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *onConflictSource[M]) String() string {
	return internal.ToString(s)
}
