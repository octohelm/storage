package sqlfrag

import (
	"context"
	contextx "github.com/octohelm/x/context"
	"iter"
	"slices"
	"strings"
)

func JoinValues(splitter string, fragments ...Fragment) Fragment {
	if len(fragments) == 0 {
		return Join(splitter, nil)
	}
	return Join(splitter, slices.Values(fragments))
}

func Join(splitter string, fragSeq iter.Seq[Fragment]) Fragment {
	return &joinFragment{
		splitter: splitter,
		fragSeq:  fragSeq,
	}
}

type joinFragment struct {
	fragSeq  iter.Seq[Fragment]
	splitter string
}

func (j *joinFragment) IsNil() bool {
	return j.fragSeq == nil

}

func (j *joinFragment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		i := 0

		for f := range NonNil(j.fragSeq) {
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
	}
}

func InlineBlock(fragment Fragment) Fragment {
	if IsNil(fragment) {
		return Empty()
	}

	return &blockFragment{
		wrapBrackets: true,
		inline:       true,
		splitter:     ",",
		fragSeq: func(yield func(Fragment) bool) {
			if !yield(fragment) {
				return
			}
		},
	}
}

func BlockWithoutBrackets(seq iter.Seq[Fragment]) Fragment {
	if seq == nil {
		return Empty()
	}

	return &blockFragment{
		splitter:     ",",
		fragSeq:      seq,
		wrapBrackets: false,
	}
}

func Block(fragment Fragment) Fragment {
	if IsNil(fragment) {
		return Empty()
	}

	return &blockFragment{
		splitter:     ",",
		wrapBrackets: true,
		fragSeq: func(yield func(Fragment) bool) {
			if !yield(fragment) {
				return
			}
		},
	}
}

type blockFragment struct {
	fragSeq      iter.Seq[Fragment]
	wrapBrackets bool
	inline       bool
	splitter     string
}

func (j *blockFragment) IsNil() bool {
	return j.fragSeq == nil
}

var identContext = contextx.New[ident]()

func (j *blockFragment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	ii := ident(0)

	if !j.inline {
		if i, ok := identContext.MayFrom(ctx); ok {
			ii = i
		}

		ii += 1
	}

	return func(yield func(string, []any) bool) {
		i := 0

		for f := range NonNil(j.fragSeq) {
			if j.wrapBrackets && i == 0 {
				if !yield("(", nil) {
					return
				}
			}

			if i > 0 {
				if !yield(j.splitter, nil) {
					return
				}
			}

			for q, args := range f.Frag(identContext.Inject(ctx, ii)) {
				if !yield(ii.tab(q), args) {
					return
				}
				i++
			}
		}

		if j.wrapBrackets && i > 0 {
			if !j.inline {
				if !yield((ii - 1).tab("\n)"), nil) {
					return
				}
				return
			}

			if !yield(")", nil) {
				return
			}
		}
	}
}

type ident int

func (i ident) tab(q string) string {
	if i > 0 && q != "" && q[0] == '\n' {
		return "\r" + strings.Repeat("\t", int(i)) + q[1:]
	}
	return q
}
