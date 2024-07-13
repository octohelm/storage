package filter

func MapFilter[A Arg, O any](args []A, where func(a Arg) (O, bool)) []O {
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
