package dal

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	contextx "github.com/octohelm/x/context"
)

func Tx(ctx context.Context, m sqlbuilder.Model, action func(ctx context.Context) error) error {
	return SessionFor(ctx, m).Tx(ctx, action)
}

var catalogs = sync.Map{}

func registerSessionCatalog(name string, tables *sqlbuilder.Tables) {
	tables.Range(func(tab sqlbuilder.Table, idx int) bool {
		catalogs.Store(tab.TableName(), name)
		return true
	})
}

type TableWrapper interface {
	Unwrap() sqlbuilder.Model
}

func SessionFor(ctx context.Context, nameOrTable any) Session {
	if u, ok := nameOrTable.(TableWrapper); ok {
		return SessionFor(ctx, u.Unwrap())
	}

	switch x := nameOrTable.(type) {
	case string:
		return FromContext(ctx, x)
	case sqlbuilder.Model:
		if t, ok := catalogs.Load(x.TableName()); ok {
			return FromContext(ctx, t.(string))
		}
	}

	return nil
}

func MustSessionFor(ctx context.Context, nameOrTable any) Session {
	s := SessionFor(ctx, nameOrTable)
	if s == nil {
		panic(errors.Errorf("invalid section target %#v", nameOrTable))
	}
	return s
}

type contextSession struct {
	name string
}

func InjectContext(ctx context.Context, repo Session) context.Context {
	return contextx.WithValue(ctx, contextSession{name: repo.Name()}, repo)
}

func FromContext(ctx context.Context, name string) Session {
	r, ok := ctx.Value(contextSession{name: name}).(Session)
	if ok {
		return r
	}
	panic(fmt.Sprintf("missing session of %s", name))
}

type Session interface {
	// Name of database
	Name() string
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
