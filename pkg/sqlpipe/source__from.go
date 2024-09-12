package sqlpipe

import (
	"context"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/x/ptr"
	"iter"
)

type FromOptionFunc[M Model] func(x *sourceFrom[M])

func FromAll[M Model](opts ...FromOptionFunc[M]) Source[M] {
	s := &sourceFrom[M]{}
	s.OptIncludesAll = ptr.Ptr(true)

	for _, o := range opts {
		o(s)
	}
	return s
}

func From[M Model](opts ...FromOptionFunc[M]) Source[M] {
	s := &sourceFrom[M]{}
	for _, o := range opts {
		o(s)
	}
	return s
}

type sourceFrom[M Model] struct {
	internal.Seed

	patchers []internal.StmtPatcher[M]
}

func (s *sourceFrom[M]) String() string {
	return internal.ToString(s)
}

func (s *sourceFrom[M]) IsNil() bool {
	return s == nil
}

func (s *sourceFrom[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *sourceFrom[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *sourceFrom[M]) ApplyStmt(ctx context.Context, b internal.StmtBuilder[M]) internal.StmtBuilder[M] {
	return internal.ApplyStmt(ctx, b.WithFlags(s.Flags), s.patchers...)
}
