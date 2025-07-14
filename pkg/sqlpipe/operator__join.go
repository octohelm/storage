package sqlpipe

import (
	"context"
	"github.com/octohelm/storage/pkg/sqltype"
	"iter"
	"strings"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func JoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		on:             on,
		from:           from,
		fromConditions: fromConditions,
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.Join(b.T(ctx, new(S)))
		},
	}
}

func JoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		on:             on,
		from:           from,
		fromConditions: fromConditions,

		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.Join(b.T(ctx, new(S)))
		},
	}
}

func FullJoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.FullJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func FullJoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.FullJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func CrossJoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.CrossJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func CrossJoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.CrossJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func InnerJoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.InnerJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func InnerJoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.InnerJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func LeftJoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.LeftJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func LeftJoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.LeftJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func RightJoinOn[M Model, S Model, T comparable](
	on modelscoped.TypedColumn[M, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, M, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.RightJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

func RightJoinOnAs[M Model, B Model, S Model, T comparable](
	on modelscoped.TypedColumn[B, T],
	from modelscoped.TypedColumn[S, T],
	fromConditions ...SourceOperator[S],
) SourceOperator[M] {
	return &joinSourcerOperator[M, B, S, T]{
		create: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			return sqlbuilder.RightJoin(b.T(ctx, new(S)))
		},
		on:             on,
		from:           from,
		fromConditions: fromConditions,
	}
}

type joinSourcerOperator[M Model, B Model, S Model, T comparable] struct {
	create         func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition
	on             modelscoped.TypedColumn[B, T]
	from           modelscoped.TypedColumn[S, T]
	fromConditions []SourceOperator[S]
}

func (j *joinSourcerOperator[M, B, S, T]) OperatorType() OperatorType {
	return OperatorJoin
}

func (j *joinSourcerOperator[M, B, S, T]) Next(from Source[M]) Source[M] {
	return &joinedSource[M]{
		Embed: Embed[M]{
			Underlying: from,
		},
		applyAsAddition: func(ctx context.Context, b *internal.Builder[M]) sqlbuilder.JoinAddition {
			where := j.on.V(sqlbuilder.EqCol(j.from))
			return j.create(ctx, b).On(j.mayPatchWhere(where))
		},
	}
}

func (j *joinSourcerOperator[M, B, S, T]) ApplyToFrom(s SourceCanPatcher[M]) {
	s.AddPatchers(internal.StmtPatcherFunc[M](func(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
		where := j.on.V(sqlbuilder.EqCol(j.from))

		return b.WithTableJoins(j.create(ctx, b).On(j.mayPatchWhere(where)))
	}))
}

func (j *joinSourcerOperator[M, B, S, T]) mayPatchWhere(where sqlfrag.Fragment) sqlfrag.Fragment {
	if len(j.fromConditions) > 0 || sqltype.HasSoftDelete[S]() {
		onSrc := From[S]().Pipe(j.fromConditions...)

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
