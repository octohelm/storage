package filter

import (
	"iter"
)

func MapFilter[T comparable, A Arg, O any](args iter.Seq[A], where func(a *Filter[T]) (O, bool)) iter.Seq[O] {
	return MapWhere[A, O](args, func(arg Arg) (O, bool) {
		switch x := arg.(type) {
		case *Filter[T]:
			return where(x)
		case Filter[T]:
			return where(&x)
		}
		return *new(O), false
	})
}

func MapWhere[A Arg, O any](args iter.Seq[A], where func(a Arg) (O, bool)) iter.Seq[O] {
	return func(yield func(O) bool) {
		for x := range args {
			v, ok := where(x)
			if ok && !yield(v) {
				return
			}
		}
	}
}

func First[A Arg, O any](args iter.Seq[A], where func(a Arg) (O, bool)) (O, bool) {
	for x := range args {
		v, ok := where(x)
		if ok {
			return v, true
		}
	}
	return *new(O), false
}
