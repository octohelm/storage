// Package session 提供数据库会话、catalog 注册与上下文注入能力。
package session

import (
	"context"

	"github.com/octohelm/storage/internal/sql/adapter"
)

// Adapter 复用底层会话适配器接口。
type Adapter = adapter.Adapter

// Open 解析 endpoint，并用已注册驱动打开会话适配器。
func Open(ctx context.Context, endpoint string) (Adapter, error) {
	return adapter.Open(ctx, endpoint)
}
