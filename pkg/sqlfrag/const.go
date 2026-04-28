package sqlfrag

import (
	"context"
	"iter"
)

// Const 表示不带绑定参数的原始 SQL 片段。
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

// Empty 返回一个不会输出任何内容的片段。
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
