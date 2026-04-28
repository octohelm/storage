package sqlfrag

import (
	"iter"
)

// NonNil 过滤掉空片段。
func NonNil[F Fragment](fragSeq iter.Seq[F]) iter.Seq[Fragment] {
	return func(yield func(Fragment) bool) {
		for frag := range fragSeq {
			if IsNil(frag) {
				continue
			}

			if !yield(frag) {
				return
			}
		}
	}
}

// Map 把输入序列映射为片段序列。
func Map[I any, O Fragment](seq iter.Seq[I], mapper func(i I) O) iter.Seq[Fragment] {
	return func(yield func(Fragment) bool) {
		for item := range seq {
			if !yield(mapper(item)) {
				return
			}
		}
	}
}
