package sqlpipe

import (
	"context"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"iter"

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
	switch x := src.(type) {
	case *sortedSource[M]:
		// self compose
		return x.withOrders(order)
	default:
		return &sortedSource[M]{
			Embed: Embed[M]{
				Underlying: src,
			},
			orders: []sqlbuilder.Order{order},
		}
	}
}

type sortedSource[M Model] struct {
	Embed[M]

	orders []sqlbuilder.Order
}

func (s *sortedSource[M]) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return internal.CollectStmt(ctx, s)
}

func (s *sortedSource[M]) ApplyStmt(ctx context.Context, b internal.StmtBuilder[M]) internal.StmtBuilder[M] {
	return s.Underlying.ApplyStmt(ctx, b.WithAdditions(
		sqlbuilder.OrderBy(s.orders...),
	))
}

func (s sortedSource[M]) withOrders(order sqlbuilder.Order) Source[M] {
	s.orders = append(s.orders, order)
	return &s
}

func (s *sortedSource[M]) Pipe(operators ...SourceOperator[M]) Source[M] {
	return Pipe[M](s, operators...)
}

func (s *sortedSource[M]) String() string {
	return internal.ToString(s)
}
