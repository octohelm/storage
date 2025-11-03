package sqlpipe

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
)

type Embed[M Model] struct {
	Underlying Source[M]
}

func (u *Embed[M]) Unwrap() Source[M] {
	return u.Underlying
}

func (u *Embed[M]) GetFlag(ctx context.Context) flags.Flag {
	src := u.Underlying

	for {
		if s, ok := src.(interface{ Unwrap() Source[M] }); ok {
			src = s.Unwrap()

			continue
		}

		break
	}

	if x, ok := src.(internal.WithFlag); ok {
		return x.GetFlag(ctx)
	}

	return 0
}

func (e *Embed[M]) IsNil() bool {
	return e.Underlying == nil || sqlfrag.IsNil(e.Underlying)
}
