package sqlite

import (
	"cmp"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"

	"modernc.org/sqlite"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/sqlfrag"
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

	query := dsn.Query()
	dbUri := dsn.Path

	connector := a.Connector()

	conn, err := connector.OpenConnector(dbUri)
	if err != nil {
		return nil, fmt.Errorf("connect failed with %s: %w", dsn.Path, err)
	}

	db := sql.OpenDB(conn)
	db.SetMaxOpenConns(1)

	adaptor := &sqliteAdapter{
		DB: adapter.Wrap(db, func(err error) error {
			if isErrorConflict(err) {
				return dberr.New(dberr.ErrTypeConflict, err.Error())
			}
			return err
		}),
	}

	journalMode := cmp.Or(query.Get("journal_mode"), "WAL")
	busyTimeout := cmp.Or(query.Get("busy_timeout"), "5000")
	synchronous := cmp.Or(query.Get("synchronous"), "NORMAL")

	if _, err := adaptor.Exec(ctx, sqlfrag.Pair(fmt.Sprintf("PRAGMA journal_mode = %s", journalMode))); err != nil {
		return nil, err
	}

	if _, err := adaptor.Exec(ctx, sqlfrag.Pair(fmt.Sprintf("PRAGMA synchronous = %s", synchronous))); err != nil {
		return nil, err
	}

	if _, err := adaptor.Exec(ctx, sqlfrag.Pair(fmt.Sprintf("PRAGMA busy_timeout = %s", busyTimeout))); err != nil {
		return nil, err
	}

	return adaptor, nil
}

func isErrorConflict(err error) bool {
	var e *sqlite.Error
	if errors.As(err, &e) && e.Code() == 2067 {
		return true
	}
	return false
}
