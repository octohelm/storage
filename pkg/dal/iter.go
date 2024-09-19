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
	return func(yield func(item *M, err error) bool) {
		chDone := make(chan error)
		defer close(chDone)

		ctx, cancel := context.WithCancel(c)
		defer func() {
			cancel()
		}()

		chItem := make(chan *M)
		go func() {
			defer close(chItem)

			chDone <- recv(ctx, func(item *M) error {
				chItem <- item
				return nil
			})
		}()

		for {
			select {
			case item := <-chItem:
				if !yield(item, nil) {
					cancel()
					// wait all done
					<-chDone
					return
				}
				continue
			case err := <-chDone:
				if err != nil {
					if !yield(nil, err) {
						return
					}
				}
				return
			}
		}
	}
}
