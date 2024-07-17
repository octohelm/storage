package compose

import (
	"context"
	"iter"
	"sync/atomic"

	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/dal/compose/querierpatcher"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type Action[M sqlbuilder.Model] struct{}

func resolveTable[T sqlbuilder.Model](ctx context.Context) sqlbuilder.Table {
	m := new(T)
	s := dal.SessionFor(ctx, m)
	return &modelTable[T]{
		Table: s.T(m),
	}
}

type modelTable[T sqlbuilder.Model] struct {
	sqlbuilder.Table
}

func (m *modelTable[T]) New() sqlbuilder.Model {
	return *new(T)
}

func (a *Action[M]) Query(patchers ...querierpatcher.Typed[M]) Result[M] {
	return newResult(func(ctx context.Context, recv Receiver[M]) error {
		return querierpatcher.ApplyTo(dal.From(resolveTable[M](ctx)), patchers...).Scan(dal.Recv(recv.Send)).Find(ctx)
	})
}

func (a *Action[M]) FindOne(ctx context.Context, patchers ...querierpatcher.Typed[M]) *M {
	i := a.Query(append(patchers, querierpatcher.Limit[M](1))...)
	for x := range i.Item(ctx) {
		return x
	}
	return nil
}

func (a *Action[M]) QueryAs(ctx context.Context, patchers ...querierpatcher.Typed[M]) dal.Querier {
	return querierpatcher.ApplyTo(dal.From(resolveTable[M](ctx)), patchers...)
}

func (a *Action[M]) CountTo(ctx context.Context, target *int64, patchers ...querierpatcher.Typed[M]) error {
	i, err := a.QueryAs(ctx, patchers...).Count(ctx)
	if err != nil {
		return err
	}
	*target = int64(i)
	return nil
}

func (a *Action[M]) List(ctx context.Context, patchers ...querierpatcher.Typed[M]) (*List[M], error) {
	list := &List[M]{
		Items: make([]*M, 0),
	}

	if err := Range(ctx, a.Query(patchers...), func(x *M) {
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
