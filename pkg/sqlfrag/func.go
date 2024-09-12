package sqlfrag

import (
	"context"
	"iter"
)

type Func func(ctx context.Context) iter.Seq2[string, []any]

func (Func) IsNil() bool {
	return false
}

func (fn Func) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return fn(ctx)
}
