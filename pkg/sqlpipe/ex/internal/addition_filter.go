package internal

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func WithoutAddition[M sqlpipe.Model](omit func(a sqlbuilder.Addition) bool) sqlpipe.SourceOperator[M] {
	return sqlpipe.SourceOperatorFunc[M](sqlpipe.OperatorSetting, func(src sqlpipe.Source[M]) sqlpipe.Source[M] {
		return &additionFilter[M]{
			Embed: sqlpipe.Embed[M]{
				Underlying: src,
			},
			omit: omit,
		}
	})
}

type additionFilter[M sqlpipe.Model] struct {
	sqlpipe.Embed[M]

	omit func(a sqlbuilder.Addition) bool
}

func (s *additionFilter[M]) Pipe(operators ...sqlpipe.SourceOperator[M]) sqlpipe.Source[M] {
	return sqlpipe.Pipe[M](s, operators...)
}

func (s *additionFilter[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (a *additionFilter[M]) ApplyStmt(ctx context.Context, s internal.StmtBuilder[M]) internal.StmtBuilder[M] {
	return a.Underlying.ApplyStmt(ctx, s).WithoutAddition(a.omit)
}
