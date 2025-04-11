package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func InsertFrom[S Model, O Model](src Source[S], cols ...modelscoped.Column[O]) Source[O] {
	return &insertFrom[S, O]{
		Embed: Embed[S]{
			Underlying: src,
		},
		cols: cols,
	}
}

type insertFrom[S Model, O Model] struct {
	Embed[S]

	cols []modelscoped.Column[O]
}

func (s *insertFrom[S, O]) IsNil() bool {
	return s.Underlying == nil
}

func (s *insertFrom[S, O]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(sqlbuilder.ContextWithToggles(ctx, sqlbuilder.Toggles{
		sqlbuilder.ToggleMultiTable: true,
	}), s)
}

func (s *insertFrom[S, O]) ApplyStmt(ctx context.Context, b *internal.Builder[O]) *internal.Builder[O] {
	selectStmt := internal.BuildStmt(ctx, s.Underlying)

	return b.WithSource(
		&internal.Mutation[O]{
			Strict: internal.Strict[O]{
				Columns: s.cols,
			},
			From: selectStmt,
		},
	)
}

func (s *insertFrom[S, O]) Pipe(operators ...SourceOperator[O]) Source[O] {
	return Pipe[O](s, operators...)
}
