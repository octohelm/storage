package session

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestOpen(t *testing.T) {
	adapter, err := Open(context.Background(), "unknown://x")
	Then(t, "Open 透传到底层 adapter.Open",
		Expect(adapter == nil, Equal(true)),
		Expect(err != nil, Equal(true)),
	)
}
