package sqlbuilder

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

const (
	ToggleMultiTable    = "MultiTable"
	ToggleNeedAutoAlias = "NeedAlias"
	ToggleUseValues     = "UseValues"
	ToggleInProject     = "InProject"
)

// Toggles 表示 SQL 构建过程中的上下文开关集合。
type Toggles map[string]bool

func (toggles Toggles) InjectContext(ctx context.Context) context.Context {
	return ContextWithToggles(ctx, toggles)
}

func (toggles Toggles) Merge(next Toggles) Toggles {
	final := Toggles{}

	for k, v := range toggles {
		if v {
			final[k] = true
		}
	}

	for k, v := range next {
		if v {
			final[k] = true
		} else {
			delete(final, k)
		}
	}

	return final
}

func (toggles Toggles) Is(key string) bool {
	if v, ok := toggles[key]; ok {
		return v
	}
	return false
}

type contextKeyForToggles struct{}

// ContextWithToggles 把开关注入 context，并与已有开关合并。
func ContextWithToggles(ctx context.Context, toggles Toggles) context.Context {
	return contextx.WithValue(ctx, contextKeyForToggles{}, TogglesFromContext(ctx).Merge(toggles))
}

// TogglesFromContext 返回 context 中的开关集合。
func TogglesFromContext(ctx context.Context) Toggles {
	if ctx == nil {
		return Toggles{}
	}
	if toggles, ok := ctx.Value(contextKeyForToggles{}).(Toggles); ok {
		return toggles
	}
	return Toggles{}
}
