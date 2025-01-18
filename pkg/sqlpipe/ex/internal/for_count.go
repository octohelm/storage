package internal

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

func ForCount[M sqlpipe.Model]() sqlpipe.SourceOperator[M] {
	return sqlpipe.SourceOperatorFunc[M](sqlpipe.OperatorSetting, func(src sqlpipe.Source[M]) sqlpipe.Source[M] {
		return &forCount[M]{
			Embed: sqlpipe.Embed[M]{
				Underlying: src,
			},
		}
	})
}

type forCount[M sqlpipe.Model] struct {
	sqlpipe.Embed[M]
}

func (s *forCount[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	return sqlpipe.Pipe[M](s, operators...)
}

func (s *forCount[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (a *forCount[M]) ApplyStmt(ctx context.Context, s *internal.Builder[M]) *internal.Builder[M] {
	return a.Underlying.ApplyStmt(ctx, s).WithFlag(flags.WithoutSorter | flags.WithoutPager)
}
