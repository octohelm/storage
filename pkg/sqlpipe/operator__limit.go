package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func Offset(offset int64) LimitOptionFunc {
	return func(o LimitOption) {
		o.SetOffset(offset)
	}
}

type LimitOptionFunc = func(o LimitOption)

type LimitOption interface {
	SetOffset(offset int64)
}

func Limit[M Model](limit int64, optFns ...LimitOptionFunc) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorLimit, func(src Source[M]) Source[M] {
		if limit < 0 {
			return src
		}

		switch x := src.(type) {
		case *limitedSource[M]:
			// self compose
			return x.with(limit, optFns...)
		}

		s := &limitedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			limit: limit,
		}
		s.build(optFns...)
		return s
	})
}

type limitedSource[M Model] struct {
	Embed[M]

	limit  int64
	offset int64
}

func (s *limitedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *limitedSource[M]) SetOffset(offset int64) {
	s.offset = offset
}

func (s *limitedSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	frag := sqlbuilder.Limit(s.limit)
	if s.offset > 0 {
		frag = frag.Offset(s.offset)
	}
	return s.Underlying.ApplyStmt(ctx, b.WithPager(frag))
}

func (s *limitedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *limitedSource[M]) String() string {
	return internal.ToString(s)
}

func (s limitedSource[M]) with(limit int64, optFns ...LimitOptionFunc) Source[M] {
	s.limit = limit
	s.build(optFns...)
	return &s
}

func (s *limitedSource[M]) build(optFns ...LimitOptionFunc) {
	for _, optFn := range optFns {
		optFn(s)
	}
}
