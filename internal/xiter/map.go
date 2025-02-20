package xiter

import "iter"

func Map[I any, O any](seq iter.Seq[I], m func(e I) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		for e := range seq {
			if !yield(m(e)) {
				return
			}
		}
	}
}
