package sqlpipe

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
)

type Embed[M Model] struct {
	Underlying Source[M]
}

func (u *Embed[M]) Unwrap() Source[M] {
	return u.Underlying
}

func (u *Embed[M]) GetFlags(ctx context.Context) internal.Flags {
	src := u.Underlying

	for {
		if s, ok := src.(interface{ Unwrap() Source[M] }); ok {
			src = s.Unwrap()

			continue
		}

		break
	}

	if x, ok := src.(internal.WithFlags); ok {
		return x.GetFlags(ctx)
	}

	return internal.Flags{}
}

func (e *Embed[M]) IsNil() bool {
	return e.Underlying == nil || sqlfrag.IsNil(e.Underlying)
}
