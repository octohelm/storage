package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func As[M Model](src Source[M], name string, opts ...FromPatcher[M]) Source[M] {
	s := &sourceAlias[M]{
		embed: Embed[M]{
			Underlying: src,
		},
		name: name,
	}

	for _, p := range opts {
		p.ApplyToFrom(s)
	}

	s.Flag = s.embed.GetFlag(context.Background())

	return s
}

type sourceAlias[M Model] struct {
	internal.Seed
	embed    Embed[M]
	name     string
	patchers []internal.StmtPatcher[M]
}

func (s *sourceAlias[M]) AddPatchers(patchers ...internal.StmtPatcher[M]) {
	s.patchers = append(s.patchers, patchers...)
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

func (s *sourceAlias[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	stmt := internal.BuildStmt(ctx, append([]internal.StmtPatcher[M]{
		s.embed.Underlying,
	}, s.patchers...)...)

	return b.WithFlag(s.GetFlag(ctx)).WithSource(
		sqlfrag.Pair("? AS ?",
			sqlfrag.Block(stmt),
			sqlfrag.Const(s.name),
		),
	)
}
