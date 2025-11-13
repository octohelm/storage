package duckdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"

	"github.com/duckdb/duckdb-go/v2"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/dberr"
)

func init() {
	adapter.Register(&duckdbAdapter{})
}

func Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	return (&duckdbAdapter{}).Open(ctx, dsn)
}

type duckdbAdapter struct {
	dialect
	adapter.DB
}

func (duckdbAdapter) DriverName() string {
	return "duckdb"
}

func (a *duckdbAdapter) Dialect() adapter.Dialect {
	return &a.dialect
}

func (a *duckdbAdapter) Connector() driver.DriverContext {
	return loggingdriver.Wrap(
		&duckdb.Driver{},
		a.DriverName(),
		func(err error) int {
			if isErrorConflict(err) {
				return 0
			}
			return 1
		},
	)
}

func (a *duckdbAdapter) Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	if a.DriverName() != dsn.Scheme {
		return nil, fmt.Errorf("invalid schema %s", dsn)
	}

	dbUri := dsn.Path

	connector := a.Connector()

	conn, err := connector.OpenConnector(dbUri)
	if err != nil {
		return nil, fmt.Errorf("connect failed with %s: %w", dsn.Path, err)
	}

	db := sql.OpenDB(conn)

	adaptor := &duckdbAdapter{
		DB: adapter.Wrap(db, func(err error) error {
			if isErrorConflict(err) {
				return dberr.New(dberr.ErrTypeConflict, err.Error())
			}
			return err
		}),
	}

	return adaptor, nil
}

func isErrorConflict(err error) bool {
	var e *duckdb.Error
	if errors.As(err, &e) && e.Type == duckdb.ErrorTypeConstraint {
		return true
	}
	return false
}
