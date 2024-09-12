package sqlpipe

import (
	"context"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"iter"
)

func As[M Model](src Source[M], name string) Source[M] {
	s := &sourceAlias[M]{
		embed: Embed[M]{
			Underlying: src,
		},
		name: name,
	}

	s.Flags = s.embed.GetFlags(context.Background())

	return s
}

type sourceAlias[M Model] struct {
	internal.Seed

	embed Embed[M]
	name  string
}

func (s *sourceAlias[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *sourceAlias[M]) IsNil() bool {
	return s == nil || sqlfrag.IsNil(s.embed.Underlying)
}

func (s *sourceAlias[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *sourceAlias[M]) String() string {
	return internal.ToString(s)
}

func (s *sourceAlias[M]) ApplyStmt(ctx context.Context, b internal.StmtBuilder[M]) internal.StmtBuilder[M] {
	stmt := internal.BuildStmt(ctx, s.embed.Underlying)

	return b.WithFlags(s.GetFlags(ctx)).WithSource(
		sqlfrag.Pair("? AS ?",
			sqlfrag.Block(stmt),
			sqlfrag.Const(s.name),
		),
	)
}
