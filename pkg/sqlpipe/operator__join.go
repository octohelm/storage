package sqlpipe

import (
	"context"
	"iter"
	"strings"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

type JoinSourceOperator[M Model] interface {
	SourceOperator[M]
	FromPatcher[M]
}

func JoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		on:            on,
		src:           from,
		srcConditions: fromConditions,

		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.Join(b.T(ctx, new(S)))
		},
	}
}

func FullJoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
	srcConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.FullJoin(b.T(ctx, new(S)))
		},
		on:            on,
		src:           src,
		srcConditions: srcConditions,
	}
}

func CrossJoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
	srcConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.CrossJoin(b.T(ctx, new(S)))
		},
		on:            on,
		src:           src,
		srcConditions: srcConditions,
	}
}

func InnerJoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
	srcConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.InnerJoin(b.T(ctx, new(S)))
		},
		on:            on,
		src:           src,
		srcConditions: srcConditions,
	}
}

func LeftJoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
	srcConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.LeftJoin(b.T(ctx, new(S)))
		},
		on:            on,
		src:           src,
		srcConditions: srcConditions,
	}
}

func RightJoinOn[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	src modelscoped.TypedColumn[S, T],
	srcConditions ...SourceOperator[S],
) JoinSourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.RightJoin(b.T(ctx, new(S)))
		},
		on:            on,
		src:           src,
		srcConditions: srcConditions,
	}
}

type joinSourcerOperator[M Model, B Model, S Model, T comparable] struct {
	create        func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition
	on            modelscoped.TypedColumn[B, T]
	src           modelscoped.TypedColumn[S, T]
	srcConditions []SourceOperator[S]
}

func (j *joinSourcerOperator[M, B, S, T]) OperatorType() OperatorType {
	return OperatorJoin
}

func (j *joinSourcerOperator[M, B, S, T]) Next(src Source[M]) Source[M] {
	return &joinedSource[M]{
		Embed: Embed[M]{
			Underlying: src,
		},
		applyAsAddition: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			where := j.on.V(sqlbuilder.EqCol(j.src))
			return j.create(ctx, b).On(j.mayPatchWhere(where))
		},
	}
}

func (j *joinSourcerOperator[M, B, S, T]) ApplyToFrom(s SourceCanPatcher[M]) {
	s.AddPatchers(internal.StmtPatcherFunc[M](func(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
		where := j.on.V(sqlbuilder.EqCol(j.src))

		return b.WithTableJoins(j.create(ctx, b).On(j.mayPatchWhere(where)))
	}))
}

func (j *joinSourcerOperator[M, B, S, T]) mayPatchWhere(where sqlfrag.Fragment) sqlfrag.Fragment {
	if len(j.srcConditions) > 0 {
		onSrc := From[S]().Pipe(j.srcConditions...)

		return sqlbuilder.And(where, sqlfrag.Func(func(ctx context.Context) iter.Seq2[string, []any] {
			return func(yield func(string, []any) bool) {
				canOmit := false

				for query, args := range onSrc.Frag(ctx) {
					if strings.HasPrefix(query, "WHERE") {
						canOmit = true
						continue
					}

					if canOmit {
						if !yield(query, args) {
							return
						}
					}
				}
			}
		}))
	}

	return where
}

type joinedSource[M Model] struct {
	Embed[M]

	applyAsAddition func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition
}

func (j *joinedSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return j.Underlying.ApplyStmt(ctx, b.WithTableJoins(j.applyAsAddition(ctx, b)))
}

func (s *joinedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *joinedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *joinedSource[M]) String() string {
	return internal.ToString(s)
}
