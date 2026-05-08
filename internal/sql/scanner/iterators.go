package scanner

import (
	"context"
	"iter"
	"reflect"

	reflectx "github.com/octohelm/x/reflect"
)

// Recv 将类型化回调适配为 ScanIterator。
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

// RecvFunc 通过 recv 推送类型化结果，并暴露为 iter.Seq2。
type RecvFunc[M any] func(ctx context.Context, recv func(v *M) error) error

func (recv RecvFunc[M]) Items(c context.Context) iter.Seq2[*M, error] {
	return func(yield func(item *M, err error) bool) {
		ctx, cancel := context.WithCancel(c)
		defer cancel()

		chItem := make(chan *M)
		chErr := make(chan error)
		go func() {
			defer close(chItem)
			defer close(chErr)

			chErr <- recv(
				ctx,
				func(item *M) error {
					chItem <- item
					return nil
				},
			)
		}()

		for {
			select {
			case item := <-chItem:
				if !yield(item, nil) {
					cancel()
					// wait all done
					<-chErr
					return
				}
				continue
			case err := <-chErr:
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

// ScanIterator 负责分配扫描目标并接收扫描结果。
type ScanIterator interface {
	// New a ptr value for scan
	New() any
	// Next For receive scanned value
	Next(v any) error
}

// ScanIteratorFor 为已有迭代器、切片或单值目标构造 ScanIterator。
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

// SliceScanIterator 会把每次扫描结果追加到目标切片中。
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

// SingleScanIterator 至多把一个扫描结果写入目标值。
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
