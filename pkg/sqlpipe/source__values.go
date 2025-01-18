package sqlpipe

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func ValueOmitZero[M Model](value *M, exclude ...modelscoped.Column[M]) Source[M] {
	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Values: func(yield func(*M) bool) {
				if !yield(value) {
					return
				}
			},
			OmitZero:        true,
			OmitZeroExclude: exclude,
		},
	}
}

func Value[M Model](value *M, cols ...modelscoped.Column[M]) Source[M] {
	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Values: func(yield func(*M) bool) {
				if !yield(value) {
					return
				}
			},
			StrictColumns: cols,
		},
	}
}

func Values[Slice ~[]*M, M Model](values Slice, cols ...modelscoped.Column[M]) Source[M] {
	if len(values) == 0 {
		return &sourceValues[M]{
			mutation: &internal.Mutation[M]{
				StrictColumns: cols,
			},
		}
	}

	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			StrictColumns: cols,
			Values:        slices.Values(values),
		},
	}
}

type sourceValues[M Model] struct {
	internal.Seed

	mutation *internal.Mutation[M]
}

func (s *sourceValues[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *sourceValues[M]) String() string {
	return internal.ToString(s)
}

func (s *sourceValues[M]) IsNil() bool {
	return s == nil
}

func (s *sourceValues[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *sourceValues[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return b.WithSource(s.mutation)
}
