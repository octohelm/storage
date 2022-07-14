package testutil

import (
	"context"
	"testing"

	"github.com/go-courier/logr"
)

func NewContext(t testing.TB) context.Context {
	t.Helper()
	ctx := context.Background()
	return logr.WithLogger(ctx, logr.StdLogger())
}
