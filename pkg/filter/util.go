package filter

func MapFilter[T comparable, A Arg, O any](args []A, where func(a *Filter[T]) (O, bool)) []O {
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

func MapWhere[A Arg, O any](args []A, where func(a Arg) (O, bool)) []O {
	ret := make([]O, 0, len(args))

	for _, x := range args {
		v, ok := where(x)
		if ok {
			ret = append(ret, v)
		}
	}

	return ret
}

func First[A Arg, O any](args []A, where func(a Arg) (O, bool)) (O, bool) {
	for _, x := range args {
		v, ok := where(x)
		if ok {
			return v, true
		}
	}
	return *new(O), false
}
