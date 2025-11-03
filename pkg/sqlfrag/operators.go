package sqlfrag

import (
	"iter"
)

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

func Map[I any, O Fragment](seq iter.Seq[I], mapper func(i I) O) iter.Seq[Fragment] {
	return func(yield func(Fragment) bool) {
		for item := range seq {
			if !yield(mapper(item)) {
				return
			}
		}
	}
}
