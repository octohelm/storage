package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

type FromPatcher[M Model] interface {
	ApplyToFrom(s SourceCanPatcher[M])
}

type SourceCanPatcher[M Model] interface {
	AddPatchers(patchers ...internal.StmtPatcher[M])
}

func FromAll[M Model](patchers ...FromPatcher[M]) Source[M] {
	s := &sourceFrom[M]{}
	s.Flag = flags.IncludesAll

	for _, patcher := range patchers {
		patcher.ApplyToFrom(s)
	}

	return s
}

func From[M Model](patchers ...FromPatcher[M]) Source[M] {
	s := &sourceFrom[M]{}

	for _, patcher := range patchers {
		patcher.ApplyToFrom(s)
	}

	return s
}

type sourceFrom[M Model] struct {
	internal.Seed

	patchers []internal.StmtPatcher[M]
}

func (s *sourceFrom[M]) AddPatchers(patchers ...internal.StmtPatcher[M]) {
	s.patchers = append(s.patchers, patchers...)
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

func (s *sourceFrom[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return internal.ApplyStmt(ctx, b.WithFlag(s.Flag), s.patchers...)
}
