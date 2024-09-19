package scanner

import (
	"context"
	"iter"
	"reflect"

	reflectx "github.com/octohelm/x/reflect"
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

type ScanIterator interface {
	// New a ptr value for scan
	New() any
	// Next For receive scanned value
	Next(v any) error
}

func ScanIteratorFor(v any) (ScanIterator, error) {
	switch x := v.(type) {
	case ScanIterator:
		return x, nil
	default:
		tpe := reflectx.Deref(reflect.TypeOf(v))

		if tpe.Kind() == reflect.Slice && tpe.Elem().Kind() != reflect.Uint8 {
			return &SliceScanIterator{
				elemType: tpe.Elem(),
				rv:       reflectx.Indirect(reflect.ValueOf(v)),
			}, nil
		}

		return &SingleScanIterator{target: v}, nil
	}
}

type SliceScanIterator struct {
	elemType reflect.Type
	rv       reflect.Value
}

func (s *SliceScanIterator) New() any {
	return reflectx.New(s.elemType).Addr().Interface()
}

func (s *SliceScanIterator) Next(v any) error {
	s.rv.Set(reflect.Append(s.rv, reflect.ValueOf(v).Elem()))
	return nil
}

type SingleScanIterator struct {
	target     any
	hasResults bool
}

func (s *SingleScanIterator) New() any {
	return s.target
}

func (s *SingleScanIterator) Next(v any) error {
	s.hasResults = true
	return nil
}

func (s *SingleScanIterator) MustHasRecord() bool {
	return s.hasResults
}
