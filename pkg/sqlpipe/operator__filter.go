package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/x/ptr"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

type Filter[M Model] interface {
	sqlfrag.Fragment
}

type FilterOp uint8

const (
	FilterOpAnd FilterOp = iota
	FilterOpOr
)

func NewWhere[M Model, T comparable](op FilterOp, col modelscoped.TypedColumn[M, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, op, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(valuer)
		})
	})
}

func WhereInSelectFrom[M Model, S Model, T comparable](col modelscoped.TypedColumn[M, T], colSelect modelscoped.TypedColumn[S, T], source Source[S]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpAnd, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
				s := source.Pipe(Select(colSelect))

				q := ""
				for query, _ := range s.Frag(internal.FlagsContext.Inject(ctx, internal.Flags{
					OptWhereRequired: ptr.Ptr(true),
				})) {
					if query != "" {
						q = query
						break
					}
				}

				if q == "" {
					return nil
				}

				return col.Fragment("# IN ?", sqlfrag.Block(s))
			})
		})
	})
}

func WhereNotInSelectFrom[M Model, S Model, T comparable](col modelscoped.TypedColumn[M, T], colSelect modelscoped.TypedColumn[S, T], source Source[S]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpAnd, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(func(col sqlbuilder.Column) sqlfrag.Fragment {
				s := source.Pipe(Select(colSelect))

				q := ""
				for query, _ := range s.Frag(internal.FlagsContext.Inject(ctx, internal.Flags{
					OptWhereRequired: ptr.Ptr(true),
				})) {
					if query != "" {
						q = query
						break
					}
				}

				if q == "" {
					return nil
				}

				return col.Fragment("# NOT IN ?", sqlfrag.Block(s))
			})
		})
	})
}

func Where[M Model, T comparable](col modelscoped.TypedColumn[M, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpAnd, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(valuer)
		})
	})
}

func OrWhere[M Model, T comparable](col modelscoped.TypedColumn[M, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpOr, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(valuer)
		})
	})
}

func CastWhere[M Model, U Model, T comparable](col modelscoped.TypedColumn[U, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpAnd, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(valuer)
		})
	})
}

func CastOrWhere[M Model, U Model, T comparable](col modelscoped.TypedColumn[U, T], valuer sqlbuilder.ColumnValuer[T]) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorFilter, func(src Source[M]) Source[M] {
		return newFilteredSource[M](src, FilterOpOr, func(ctx context.Context) sqlfrag.Fragment {
			return col.V(valuer)
		})
	})
}

func newFilteredSource[M Model](src Source[M], op FilterOp, builder func(ctx context.Context) sqlfrag.Fragment) Source[M] {
	switch x := src.(type) {
	case *filteredSource[M]:
		// self compose
		return x.with(op, builder)
	default:
		return &filteredSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			whereBuilders: []*whereBuilder{
				{
					op: op,
					b:  builder,
				},
			},
		}
	}
}

type filteredSource[M Model] struct {
	Embed[M]

	whereBuilders []*whereBuilder
}

type whereBuilder struct {
	op FilterOp
	b  func(ctx context.Context) sqlfrag.Fragment
}

func (s *filteredSource[M]) ApplyStmt(ctx context.Context, b internal.StmtBuilder[M]) internal.StmtBuilder[M] {
	var w sqlfrag.Fragment

	for _, b := range s.whereBuilders {
		if ww := b.b(ctx); !sqlfrag.IsNil(ww) {
			switch b.op {
			case FilterOpOr:
				if w == nil {
					w = ww
				} else {
					w = sqlbuilder.Or(w, ww)
				}
			case FilterOpAnd:
				if w == nil {
					w = ww
				} else {
					w = sqlbuilder.And(w, ww)
				}
			}
		}
	}

	if sqlfrag.IsNil(w) {
		return s.Underlying.ApplyStmt(ctx, b)
	}

	return s.Underlying.ApplyStmt(ctx, b.WithAdditions(
		sqlbuilder.Where(w),
	))
}

func (s filteredSource[M]) with(op FilterOp, b func(ctx context.Context) sqlfrag.Fragment) *filteredSource[M] {
	s.whereBuilders = append(s.whereBuilders, &whereBuilder{
		op: op,
		b:  b,
	})

	return &s
}

func (s *filteredSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *filteredSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *filteredSource[M]) String() string {
	return internal.ToString(s)
}
