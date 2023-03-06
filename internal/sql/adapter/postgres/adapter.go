package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/pkg/errors"
)

func init() {
	adapter.Register(&pgAdapter{}, "postgresql")
}

func Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	return (&pgAdapter{}).Open(ctx, dsn)
}

type pgAdapter struct {
	dialect
	adapter.DB
	dbName string
}

func (a *pgAdapter) Dialect() adapter.Dialect {
	return &a.dialect
}

func (pgAdapter) DriverName() string {
	return "postgres"
}

func (a *pgAdapter) Connector() driver.DriverContext {
	return loggingdriver.Wrap(&stdlib.Driver{}, a.DriverName(), func(err error) int {
		if pqerr, ok := dberr.UnwrapAll(err).(*pgconn.PgError); ok {
			// unique_violation
			if pqerr.Code == "23505" {
				return 0
			}
		}
		return 1
	})
}

func dbNameFromDSN(dsn *url.URL) string {
	return strings.TrimLeft(dsn.Path, "/")
}

func (a *pgAdapter) Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	if a.DriverName() != dsn.Scheme {
		return nil, errors.Errorf("invalid schema %s", dsn)
	}

	dbName := dbNameFromDSN(dsn)

	c, err := a.Connector().OpenConnector(dsn.String())
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(c)

	if err := db.PingContext(ctx); err != nil {
		if isErrorUnknownDatabase(err) {
			if err := a.createDatabase(ctx, dbName, *dsn); err != nil {
				return nil, err
			}
			return a.Open(ctx, dsn)
		}
		return nil, err
	}

	return &pgAdapter{
		dbName: dbName,
		DB: adapter.Wrap(db, func(err error) error {
			if isErrorConflict(err) {
				return dberr.New(dberr.ErrTypeConflict, err.Error())
			}
			return err
		}),
	}, nil
}

func isErrorConflict(err error) bool {
	if e, ok := dberr.UnwrapAll(err).(*pgconn.PgError); ok {
		if e.Code == "23505" {
			return true
		}
	}
	return false
}

func isErrorUnknownDatabase(err error) bool {
	if e, ok := dberr.UnwrapAll(err).(*pgconn.PgError); ok {
		if e.Code == "3D000" {
			return true
		}
	}
	return false
}

func (a *pgAdapter) createDatabase(ctx context.Context, dbName string, dsn url.URL) error {
	dsn.Path = ""

	adaptor, err := a.Open(ctx, &dsn)
	if err != nil {
		return err
	}
	defer adaptor.Close()

	_, err = adaptor.Exec(context.Background(), sqlbuilder.Expr(fmt.Sprintf("CREATE DATABASE %s", dbName)))
	return err
}
