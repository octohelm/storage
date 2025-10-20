package session

import (
	"context"

	"github.com/octohelm/storage/internal/sql/adapter"
)

type Adapter = adapter.Adapter

func Open(ctx context.Context, endpoint string) (Adapter, error) {
	return adapter.Open(ctx, endpoint)
}
