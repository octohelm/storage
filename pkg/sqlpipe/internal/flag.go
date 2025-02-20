package internal

import (
	"context"

	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
	contextx "github.com/octohelm/x/context"
)

var FlagContext = contextx.New[flags.Flag]()

type WithFlag interface {
	GetFlag(ctx context.Context) flags.Flag
}

var _ WithFlag = Seed{}

type Seed struct {
	flags.Flag
}

func (s Seed) GetFlag(ctx context.Context) flags.Flag {
	if f, ok := FlagContext.MayFrom(ctx); ok {
		return s.Flag | f
	}
	return s.Flag
}
