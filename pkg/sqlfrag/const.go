package sqlfrag

import (
	"context"
	"iter"
)

type Const string

var _ Fragment = Const("")

func (v Const) IsNil() bool {
	return len(v) == 0
}

func (v Const) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		if !yield(string(v), nil) {
			return
		}
	}
}

func Empty() Fragment {
	return &empty{}
}

type empty struct{}

func (empty) IsNil() bool {
	return true
}

func (empty) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return func(yield func(string, []any) bool) {
		return
	}
}
