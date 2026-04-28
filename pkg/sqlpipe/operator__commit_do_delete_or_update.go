package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

// DoDelete 把数据源转换为软删除提交操作。
func DoDelete[M Model]() SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		return &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForDelete: internal.DeleteTypeSoft,
			},
		}
	})
}

// DoDeleteHard 把数据源转换为硬删除提交操作。
func DoDeleteHard[M Model]() SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		return &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForDelete: internal.DeleteTypeHard,
			},
		}
	})
}

// DoUpdate 为指定列追加更新赋值。
func DoUpdate[M Model, T any](col modelscoped.TypedColumn[M, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		switch x := src.(type) {
		case *updateOrDeleteSource[M]:
			// self compose
			return x.withAssignments(col.By(valuer))
		}

		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate: true,
				Assignments: []sqlbuilder.Assignment{
					col.By(valuer),
				},
			},
		}
		return s
	})
}

// DoUpdateSet 按给定模型值和列集合构造更新。
func DoUpdateSet[M Model](m *M, columns ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate: true,
				Strict: internal.Strict[M]{
					Columns: columns,
				},
				Values: func(yield func(*M) bool) {
					if !yield(m) {
						return
					}
				},
			},
		}
		return s
	})
}

// DoUpdateSetOmit 按给定模型值构造更新，并忽略指定列。
func DoUpdateSetOmit[M Model](m *M, columns ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate: true,
				Strict: internal.Strict[M]{
					Columns: columns,
					Omit:    true,
				},
				Values: func(yield func(*M) bool) {
					if !yield(m) {
						return
					}
				},
			},
		}
		return s
	})
}

// DoUpdateSetOmitZero 按给定模型值构造更新，并忽略零值字段。
func DoUpdateSetOmitZero[M Model](m *M, exclude ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate: true,
				OmitZero: internal.OmitZero[M]{
					Enabled: true,
					Exclude: exclude,
				},
				Values: func(yield func(*M) bool) {
					yield(m)
				},
			},
		}
		return s
	})
}

type updateOrDeleteSource[M Model] struct {
	Embed[M]

	mutation *internal.Mutation[M]
}

func (s *updateOrDeleteSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *updateOrDeleteSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return s.Underlying.ApplyStmt(ctx, b.WithSource(s.mutation))
}

func (s *updateOrDeleteSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *updateOrDeleteSource[M]) String() string {
	return internal.ToString(s)
}

func (s updateOrDeleteSource[M]) withAssignments(assignments ...sqlbuilder.Assignment) Source[M] {
	s.mutation = s.mutation.WithAssignments(assignments...)
	return &s
}
