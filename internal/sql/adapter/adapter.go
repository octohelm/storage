package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

type DB interface {
	Exec(ctx context.Context, expr sqlfrag.Fragment) (sql.Result, error)
	Query(ctx context.Context, expr sqlfrag.Fragment) (*sql.Rows, error)
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
	CreateTableIsNotExists(t sqlbuilder.Table) []sqlfrag.Fragment
	DropTable(t sqlbuilder.Table) sqlfrag.Fragment
	TruncateTable(t sqlbuilder.Table) sqlfrag.Fragment

	AddColumn(col sqlbuilder.Column) sqlfrag.Fragment
	RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlfrag.Fragment
	ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlfrag.Fragment
	DropColumn(col sqlbuilder.Column) sqlfrag.Fragment

	AddIndex(key sqlbuilder.Key) sqlfrag.Fragment
	DropIndex(key sqlbuilder.Key) sqlfrag.Fragment

	DataType(columnDef sqlbuilder.ColumnDef) sqlfrag.Fragment
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
		return nil, fmt.Errorf("missing adapter for %s", u.Scheme)
	}

	return
}
