package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"

	"github.com/octohelm/storage/pkg/dberr"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/pkg/errors"
	"modernc.org/sqlite"
)

func init() {
	adapter.Register(&sqliteAdapter{})
}

func Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	return (&sqliteAdapter{}).Open(ctx, dsn)
}

type sqliteAdapter struct {
	dialect
	adapter.DB
}

func (sqliteAdapter) DriverName() string {
	return "sqlite"
}

func (a *sqliteAdapter) Dialect() adapter.Dialect {
	return &a.dialect
}

func (a *sqliteAdapter) Connector() driver.DriverContext {
	return loggingdriver.Wrap(&sqlite.Driver{}, a.DriverName(), func(err error) int {
		if e, ok := dberr.UnwrapAll(err).(*sqlite.Error); ok {
			if e.Code() == 2067 {
				return 0
			}
		}
		return 1
	})
}

func (a *sqliteAdapter) Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	if a.DriverName() != dsn.Scheme {
		return nil, errors.Errorf("invalid schema %s", dsn)
	}

	c, err := a.Connector().OpenConnector(dsn.Path)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(c)

	return &sqliteAdapter{
		DB: adapter.Wrap(db, func(err error) error {
			if isErrorConflict(err) {
				return dberr.New(dberr.ErrTypeConflict, err.Error())
			}
			return err
		}),
	}, nil
}

func isErrorConflict(err error) bool {
	if e, ok := dberr.UnwrapAll(err).(*sqlite.Error); ok && e.Code() == 2067 {
		return true
	}
	return false
}

func (a *sqliteAdapter) createDatabase(ctx context.Context, dbName string, dsn url.URL) error {
	dsn.Path = ""

	adaptor, err := a.Open(ctx, &dsn)
	if err != nil {
		return err
	}
	defer adaptor.Close()

	_, err = adaptor.Exec(context.Background(), sqlbuilder.Expr(fmt.Sprintf("CREATE DATABASE %s", dbName)))
	return err
}
