package sqlfrag

import (
	"context"
	"iter"
)

// ContextInjector 定义片段上下文注入能力。
type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
}

// WithContextInjector 为片段包装上下文注入器。
func WithContextInjector(injector ContextInjector, fragment Fragment) Fragment {
	return &contextInjectorFragment{
		injector: injector,
		fragment: fragment,
	}
}

type contextInjectorFragment struct {
	injector ContextInjector
	fragment Fragment
}

func (c *contextInjectorFragment) IsNil() bool {
	return IsNil(c.fragment)
}

func (c contextInjectorFragment) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return c.fragment.Frag(c.injector.InjectContext(ctx))
}
