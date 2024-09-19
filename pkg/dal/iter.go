package dal

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
