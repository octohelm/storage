package dal

import (
	"context"
	"iter"
)

func Recv[T any](next func(v *T) error) ScanIterator {
	return &typedScanner[T]{next: next}
}

type typedScanner[T any] struct {
	next func(v *T) error
}

func (*typedScanner[T]) New() any {
	return new(T)
}

func (t *typedScanner[T]) Next(v any) error {
	return t.next(v.(*T))
}

type RecvFunc[M any] func(ctx context.Context, recv func(v *M) error) error

func (recv RecvFunc[M]) Item(c context.Context) iter.Seq2[*M, error] {
	var err error

	chComplete := make(chan error)
	chItem := make(chan *M)

	ctx, cancel := context.WithCancel(c)
	go func() {
		e := recv(ctx, func(item *M) error {
			chItem <- item
			return nil
		})

		close(chItem)
		err = e
		chComplete <- e
	}()

	return func(yield func(item *M, err error) bool) {
		defer func() {
			cancel()

			// wait complete to avoid db deadlock
			<-chComplete
			close(chComplete)
		}()

		for item := range chItem {
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
