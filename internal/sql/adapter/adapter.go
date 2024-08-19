package adapter

import (
	"context"
	"database/sql"
	"net/url"
	"sync"

	"github.com/pkg/errors"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type DB interface {
	Exec(ctx context.Context, expr sqlbuilder.SqlExpr) (sql.Result, error)
	Query(ctx context.Context, expr sqlbuilder.SqlExpr) (*sql.Rows, error)
	Transaction(ctx context.Context, action func(ctx context.Context) error) error
	Close() error
}

type Connector interface {
	Open(ctx context.Context, dsn *url.URL) (Adapter, error)
}

type Adapter interface {
	DB
	DriverName() string
	Dialect() Dialect
	Catalog(ctx context.Context) (*sqlbuilder.Tables, error)
}

type Dialect interface {
	CreateTableIsNotExists(t sqlbuilder.Table) []sqlbuilder.SqlExpr
	DropTable(t sqlbuilder.Table) sqlbuilder.SqlExpr
	TruncateTable(t sqlbuilder.Table) sqlbuilder.SqlExpr

	AddColumn(col sqlbuilder.Column) sqlbuilder.SqlExpr
	RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlbuilder.SqlExpr
	ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlbuilder.SqlExpr
	DropColumn(col sqlbuilder.Column) sqlbuilder.SqlExpr

	AddIndex(key sqlbuilder.Key) sqlbuilder.SqlExpr
	DropIndex(key sqlbuilder.Key) sqlbuilder.SqlExpr

	DataType(columnDef sqlbuilder.ColumnDef) sqlbuilder.SqlExpr
}

var adapters = sync.Map{}

func Register(a Adapter, aliases ...string) {
	adapters.Store(a.DriverName(), a)
	for i := range aliases {
		adapters.Store(aliases[i], a)
	}
}

func Open(ctx context.Context, dsn string) (a Adapter, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	adapters.Range(func(key, value any) bool {
		if key.(string) == u.Scheme {
			a, err = value.(Connector).Open(ctx, u)
			return false
		}
		return true
	})

	if a == nil && err == nil {
		return nil, errors.Errorf("missing adapter for %s", u.Scheme)
	}

	return
}
