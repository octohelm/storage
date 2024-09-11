package sqlfrag

import (
	"context"
	"iter"
)

type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
}

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
