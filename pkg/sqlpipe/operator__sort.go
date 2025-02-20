package sqlpipe

import (
	"context"
	"iter"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

func CastAscSort[M Model, U Model](col modelscoped.Column[U], ex ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorSort, func(src Source[M]) Source[M] {
		return newSortedSource(src, sqlbuilder.AscOrder(col, ex...))
	})
}

func CastDescSort[M Model, U Model](col modelscoped.Column[U], ex ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorSort, func(src Source[M]) Source[M] {
		return newSortedSource(src, sqlbuilder.DescOrder(col, ex...))
	})
}

func AscSort[M Model](col modelscoped.Column[M], ex ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorSort, func(src Source[M]) Source[M] {
		return newSortedSource(src, sqlbuilder.AscOrder(col, ex...))
	})
}

func DescSort[M Model](col modelscoped.Column[M], ex ...sqlfrag.Fragment) SourceOperator[M] {
	return SourceOperatorFunc[M](OperatorSort, func(src Source[M]) Source[M] {
		return newSortedSource(src, sqlbuilder.DescOrder(col, ex...))
	})
}

func newSortedSource[M Model](src Source[M], order sqlbuilder.Order) Source[M] {
	return &sortedSource[M]{
		Embed: Embed[M]{
			Underlying: src,
		},
		order: order,
	}
}

type sortedSource[M Model] struct {
	Embed[M]
	order sqlbuilder.Order
}

func (s *sortedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *sortedSource[M]) ApplyStmt(ctx context.Context, b *internal.Builder[M]) *internal.Builder[M] {
	return s.Underlying.ApplyStmt(ctx, b.WithOrders(s.order))
}

func (s *sortedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *sortedSource[M]) String() string {
	return internal.ToString(s)
}
