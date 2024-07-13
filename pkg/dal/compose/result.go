package compose

import (
	"context"
	"iter"
)

func Range[T any](ctx context.Context, ret Result[T], do func(x *T)) error {
	for item := range ret.Item(ctx) {
		do(item)
	}
	return ret.Err()
}

type Result[T any] interface {
	Item(ctx context.Context) iter.Seq[*T]
	Err() error
}
