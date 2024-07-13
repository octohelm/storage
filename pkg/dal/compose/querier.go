package compose

import (
	"context"
	"iter"
	"sync/atomic"

	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/dal/compose/querierpatcher"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type Querier[M sqlbuilder.Model] struct{}

func (q *Querier[M]) QueryAs(ctx context.Context, patchers ...querierpatcher.Typed[M]) dal.Querier {
	m := new(M)
	s := dal.SessionFor(ctx, m)
	return querierpatcher.ApplyTo(dal.From(s.T(m)), patchers...)
}

func (q *Querier[M]) Query(patchers ...querierpatcher.Typed[M]) Result[M] {
	m := new(M)

	return newResult(func(ctx context.Context, recv Receiver[M]) error {
		s := dal.SessionFor(ctx, m)

		return querierpatcher.ApplyTo(dal.From(s.T(m)), patchers...).Scan(dal.Recv(recv.Send)).Find(ctx)
	})
}

func (q *Querier[M]) FindOne(ctx context.Context, patchers ...querierpatcher.Typed[M]) *M {
	i := q.Query(append(patchers, querierpatcher.Limit[M](1))...)
	for x := range i.Item(ctx) {
		return x
	}
	return nil
}
func (q *Querier[M]) Count(ctx context.Context, patchers ...querierpatcher.Typed[M]) (int, error) {
	return q.QueryAs(ctx, patchers...).Count(ctx)
}

func (q *Querier[M]) List(ctx context.Context, patchers ...querierpatcher.Typed[M]) (*List[M], error) {
	list := &List[M]{
		Items: make([]*M, 0),
	}

	if err := Range(ctx, q.Query(patchers...), func(x *M) {
		list.Items = append(list.Items, x)
	}); err != nil {
		return nil, err
	}

	return list, nil
}

type List[T any] struct {
	Items []*T `json:"items"`
}

type Receiver[T any] interface {
	Send(t *T) error
}

func newResult[T any](query func(ctx context.Context, recv Receiver[T]) error) Result[T] {
	return &rowIter[T]{
		query: query,
	}
}

type rowIter[T any] struct {
	query func(ctx context.Context, recv Receiver[T]) error
	err   atomic.Value
}

func (rt *rowIter[T]) Item(ctx context.Context) iter.Seq[*T] {
	return func(yield func(item *T) bool) {
		c, cancel := context.WithCancel(ctx)
		defer cancel()

		ch := make(chan *T)

		go func() {
			defer close(ch)

			if err := rt.query(c, &recv[T]{ctx: ctx, ch: ch}); err != nil {
				rt.err.Store(err)
			}
		}()

		for item := range ch {
			if !yield(item) {
				return
			}
		}
	}
}

func (rt *rowIter[T]) Err() error {
	if err, ok := rt.err.Load().(error); ok {
		if err == nil {
			return nil
		}
		return err
	}
	return nil
}

type recv[T any] struct {
	ch  chan *T
	ctx context.Context
}

func (r *recv[T]) Send(item *T) error {
	select {
	case <-r.ctx.Done():
		return r.ctx.Err()
	case r.ch <- item:
	}
	return nil
}
