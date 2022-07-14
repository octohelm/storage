package dal

import (
	"context"
	"fmt"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	contextx "github.com/octohelm/x/context"
)

type repositoryContext struct {
	name string
}

func InjectContext(ctx context.Context, repo Session) context.Context {
	return contextx.WithValue(ctx, repositoryContext{name: repo.Name()}, repo)
}

func FromContext(ctx context.Context, name string) Session {
	r, ok := ctx.Value(repositoryContext{name: name}).(Session)
	if ok {
		return r
	}
	panic(fmt.Sprintf("missing session of %s", name))
}

type Session interface {
	// Name of database
	Name() string
	Close() error
	T(m any) sqlbuilder.Table
	Tx(ctx context.Context, fn func(ctx context.Context) error) error
	Adapter() adapter.Adapter
}

func New(a adapter.Adapter, name string) Session {
	return &session{
		name:    name,
		adapter: a,
	}
}

type session struct {
	name    string
	adapter adapter.Adapter
}

func (s *session) Close() error {
	return s.adapter.Close()
}

func (s *session) Adapter() adapter.Adapter {
	return s.adapter
}

func (s *session) Name() string {
	return s.name
}

func (s *session) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.adapter.Transaction(ctx, fn)
}

func (s *session) T(m any) sqlbuilder.Table {
	if td, ok := m.(sqlbuilder.TableDefinition); ok {
		return td.T()
	}
	if td, ok := m.(sqlbuilder.Table); ok {
		return td
	}
	return sqlbuilder.TableFromModel(m)
}
