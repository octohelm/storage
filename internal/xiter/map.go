package xiter

import (
	"iter"
)

// Map 把输入序列逐项映射为新序列。
func Map[I any, O any](seq iter.Seq[I], m func(e I) O) iter.Seq[O] {
	return func(yield func(O) bool) {
		for e := range seq {
			if !yield(m(e)) {
				return
			}
		}
	}
}
