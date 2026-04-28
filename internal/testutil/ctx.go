package testutil

import (
	"context"
	"testing"

	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
)

// NewContext 创建带默认日志器的测试上下文。
func NewContext(t testing.TB) context.Context {
	t.Helper()
	ctx := context.Background()
	return logr.WithLogger(ctx, slog.Logger(slog.Default()))
}
