package compose

import (
	"context"
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/dal/compose/patcher"
	dalcomposetarget "github.com/octohelm/storage/pkg/dal/compose/target"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"iter"
	"sync"
)

type Action[M sqlbuilder.Model] struct{}

func (a *Action[M]) Query(patchers ...patcher.TypedQuerierPatcher[M]) Result[M] {
	return recvFunc[M](func(ctx context.Context, recv func(v *M) error) error {
		q := dal.From(dalcomposetarget.Table[M](ctx))

		return patcher.ApplyToQuerier(q, patchers...).
			Scan(dal.Recv(recv)).
			Find(ctx)
	})
}

func (a *Action[M]) QueryAs(ctx context.Context, patchers ...patcher.TypedQuerierPatcher[M]) dal.Querier {
	q := dal.From(dalcomposetarget.Table[M](ctx))

	return patcher.ApplyToQuerier(q, patchers...)
}

func (a *Action[M]) FindOne(ctx context.Context, patchers ...patcher.TypedQuerierPatcher[M]) *M {
	i := a.Query(append(patchers, patcher.Limit[M](1))...)
	for x, _ := range i.Item(ctx) {
		return x
	}
	return nil
}

func (a *Action[M]) CountTo(ctx context.Context, target *int64, patchers ...patcher.TypedQuerierPatcher[M]) error {
	i, err := a.QueryAs(ctx, patchers...).Count(ctx)
	if err != nil {
		return err
	}
	*target = int64(i)
	return nil
}

func (a *Action[M]) List(ctx context.Context, patchers ...patcher.TypedQuerierPatcher[M]) (*List[M], error) {
	ret := a.Query(patchers...)

	return ToList(ctx, ret)
}

func (a *Action[M]) Mutate(ctx context.Context, m dal.Mutation[M], patchers ...dal.MutationPatcher[M]) error {
	return m.Apply(patchers...).Save(ctx)
}

func (a *Action[M]) MutateThenQuery(m dal.Mutation[M], patchers ...dal.MutationPatcher[M]) Result[M] {
	return recvFunc[M](func(ctx context.Context, next func(v *M) error) error {
		return patcher.ApplyToMutation(m, patchers...).
			Scan(dal.Recv(next)).
			Save(ctx)
	})
}

func (a *Action[M]) MutateThenList(ctx context.Context, m dal.Mutation[M], patchers ...dal.MutationPatcher[M]) (*List[M], error) {
	ret := a.MutateThenQuery(m, patchers...)
	return ToList(ctx, ret)
}

func Range[T any](ctx context.Context, ret Result[T], yield func(x *T)) error {
	for item, err := range ret.Item(ctx) {
		if err != nil {
			return err
		}
		yield(item)
	}
	return nil
}

type Result[T any] interface {
	Item(ctx context.Context) iter.Seq2[*T, error]
}

func ToList[M any](ctx context.Context, ret Result[M]) (*List[M], error) {
	list := &List[M]{
		Items: make([]*M, 0),
	}

	for x, err := range ret.Item(ctx) {
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, x)
	}

	return list, nil
}

type List[M any] struct {
	Items []*M `json:"items"`
}

type recvFunc[M any] func(ctx context.Context, recv func(v *M) error) error

func (recv recvFunc[M]) Item(ctx context.Context) iter.Seq2[*M, error] {
	once := &sync.Once{}
	var err error
	ch := make(chan *M)

	done := func(e error) {
		once.Do(func() {
			close(ch)
			err = e
		})
	}

	go func() {
		done(recv(ctx, func(item *M) error {
			ch <- item
			return nil
		}))
	}()

	return func(yield func(item *M, err error) bool) {
		defer done(nil)

		for item := range ch {
			if !yield(item, nil) {
				return
			}
		}

		if err != nil {
			if !yield(nil, err) {
				return
			}
		}
	}
}
