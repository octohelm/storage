package sqlpipe

import (
	"context"
	"iter"
	"slices"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

// ValueOmitZero 插入单个值，并忽略零值字段，显式排除列除外。
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

// Value 插入单个值，并且只使用指定列。
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

// ValueOmit 插入单个值，并忽略指定列。
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

// ValueSeq 收集迭代器中的值，并按严格列集合插入。
func ValueSeq[M Model](values iter.Seq[*M], strictColumns ...modelscoped.Column[M]) Source[M] {
	return Values(slices.Collect(values), strictColumns...)
}

// Values 插入一个值切片，并可指定严格列集合。
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

// ValueSeqOmit 收集迭代器中的值，并在插入时忽略指定列。
func ValueSeqOmit[M Model](values iter.Seq[*M], columns ...modelscoped.Column[M]) Source[M] {
	return ValuesOmit(slices.Collect(values), columns...)
}

// ValuesOmit 插入一个值切片，并忽略指定列。
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
