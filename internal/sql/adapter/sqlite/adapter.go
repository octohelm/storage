package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"

	"modernc.org/sqlite"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/dberr"
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
	mutexSet
}

func (sqliteAdapter) DriverName() string {
	return "sqlite"
}

func (a *sqliteAdapter) Dialect() adapter.Dialect {
	return &a.dialect
}

func (a *sqliteAdapter) Connector() driver.DriverContext {
	return loggingdriver.Wrap(
		&sqlite.Driver{},
		a.DriverName(),
		func(err error) int {
			var e *sqlite.Error
			if errors.As(err, &e) {
				if e.Code() == 2067 {
					return 0
				}
			}
			return 1
		},
	)
}

func (a *sqliteAdapter) Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	if a.DriverName() != dsn.Scheme {
		return nil, fmt.Errorf("invalid schema %s", dsn)
	}

	dbUri := dsn.Path + "?" + dsn.Query().Encode()

	connector := &driverContextWithMutex{
		DriverContext: a.Connector(),
		Mutex:         a.of(dbUri),
	}

	conn, err := connector.OpenConnector(dbUri)
	if err != nil {
		return nil, fmt.Errorf("connect failed with %s: %w", dsn.Path, err)
	}

	db := sql.OpenDB(conn)

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
	var e *sqlite.Error
	if errors.As(err, &e) && e.Code() == 2067 {
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

	_, err = adaptor.Exec(context.Background(), sqlfrag.Pair(fmt.Sprintf("CREATE DATABASE %s", dbName)))
	return err
}
