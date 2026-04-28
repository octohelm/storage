package internal

import (
	"context"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

// FlagContext 在上下文中传递 sqlpipe 标记位。
var FlagContext = contextx.New[flags.Flag]()

// WithFlag 表示对象可从上下文解析并合并标记位。
type WithFlag interface {
	GetFlag(ctx context.Context) flags.Flag
}

var _ WithFlag = Seed{}

// Seed 提供基础标记位实现。
type Seed struct {
	flags.Flag
}

func (s Seed) GetFlag(ctx context.Context) flags.Flag {
	if f, ok := FlagContext.MayFrom(ctx); ok {
		return s.Flag | f
	}
	return s.Flag
}
