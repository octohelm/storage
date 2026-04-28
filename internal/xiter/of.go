package xiter

import (
	"iter"
)

// Of 把给定值列表包装为 iter.Seq。
func Of[T any](values ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}
