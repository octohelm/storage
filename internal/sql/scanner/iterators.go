package scanner

import (
	"reflect"

	reflectx "github.com/octohelm/x/reflect"
)

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
