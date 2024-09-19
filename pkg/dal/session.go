package dal

import (
	"context"

	"github.com/octohelm/storage/pkg/session"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Tx(ctx context.Context, m sqlbuilder.Model, action func(ctx context.Context) error) error {
	return session.For(ctx, m).Tx(ctx, action)
}

func SessionFor(ctx context.Context, m any) Session {
	return session.For(ctx, m)
}

type Session = session.Session
