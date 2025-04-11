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
			OmitZero: internal.OmitZero[M]{
				Enabled: true,
				Exclude: exclude,
			},
		},
	}
}

func Value[M Model](value *M, columns ...modelscoped.Column[M]) Source[M] {
	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Values: func(yield func(*M) bool) {
				if !yield(value) {
					return
				}
			},
			Strict: internal.Strict[M]{
				Columns: columns,
			},
		},
	}
}

func ValueOmit[M Model](value *M, columns ...modelscoped.Column[M]) Source[M] {
	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Values: func(yield func(*M) bool) {
				if !yield(value) {
					return
				}
			},
			Strict: internal.Strict[M]{
				Columns: columns,
				Omit:    true,
			},
		},
	}
}

func ValueSeq[M Model](values iter.Seq[*M], strictColumns ...modelscoped.Column[M]) Source[M] {
	return Values(slices.Collect(values), strictColumns...)
}

func Values[Slice ~[]*M, M Model](values Slice, strictColumns ...modelscoped.Column[M]) Source[M] {
	if len(values) == 0 {
		return &noop[M]{}
	}

	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Strict: internal.Strict[M]{
				Columns: strictColumns,
			},
			Values: slices.Values(values),
		},
	}
}

func ValueSeqOmit[M Model](values iter.Seq[*M], columns ...modelscoped.Column[M]) Source[M] {
	return ValuesOmit(slices.Collect(values), columns...)
}

func ValuesOmit[Slice ~[]*M, M Model](values Slice, columns ...modelscoped.Column[M]) Source[M] {
	if len(values) == 0 {
		return &noop[M]{}
	}

	return &sourceValues[M]{
		mutation: &internal.Mutation[M]{
			Strict: internal.Strict[M]{
				Omit:    true,
				Columns: columns,
			},
			Values: slices.Values(values),
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

type noop[M Model] struct{}

func (*noop[M]) IsNil() bool {
	return true
}

func (n *noop[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
	}
}

func (n *noop[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return n
}

func (n *noop[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return b
}
