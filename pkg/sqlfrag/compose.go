package sqlfrag

import (
	"context"
	"iter"
	"slices"
)

func Group(fragment Fragment) Fragment {
	if IsNil(fragment) {
		return Empty()
	}

	return &groupFragment{
		group:    true,
		splitter: ",",
		fragSeq: func(yield func(Fragment) bool) {
			if !yield(fragment) {
				return
			}
		},
	}
}

func JoinValues(splitter string, fragments ...Fragment) Fragment {
	if len(fragments) == 0 {
		return Join(splitter, nil)
	}
	return Join(splitter, slices.Values(fragments))
}

func Join(splitter string, fragSeq iter.Seq[Fragment]) Fragment {
	return &groupFragment{
		splitter: splitter,
		fragSeq:  fragSeq,
	}
}

type groupFragment struct {
	group    bool
	splitter string
	fragSeq  iter.Seq[Fragment]
}

func (j *groupFragment) IsNil() bool {
	return j.fragSeq == nil
}

func (j *groupFragment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		i := 0

		for f := range NonNil(j.fragSeq) {
			if j.group && i == 0 {
				if !yield("(", nil) {
					return
				}
			}

			if i > 0 {
				if !yield(j.splitter, nil) {
					return
				}
			}

			for q, args := range f.Frag(ctx) {
				if !yield(q, args) {
					return
				}
				i++
			}
		}

		if j.group && i > 0 {
			if !yield(")", nil) {
				return
			}
		}
	}
}
