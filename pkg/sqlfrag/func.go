package sqlfrag

import (
	"context"
	"iter"
)

// Func 允许直接用函数实现 Fragment。
type Func func(ctx context.Context) iter.Seq2[string, []any]

func (Func) IsNil() bool {
	return false
}

func (fn Func) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return fn(ctx)
}
