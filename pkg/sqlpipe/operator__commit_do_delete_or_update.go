package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"

	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

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

func DoUpdateSet[M Model](m *M, cols ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate:     true,
				StrictColumns: cols,
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

func DoUpdateSetOmitZero[M Model](m *M, exclude ...modelscoped.Column[M]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorCommit, func(src Source[M]) Source[M] {
		s := &updateOrDeleteSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			mutation: &internal.Mutation[M]{
				ForUpdate:       true,
				OmitZero:        true,
				OmitZeroExclude: exclude,
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
