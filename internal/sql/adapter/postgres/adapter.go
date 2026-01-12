package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/loggingdriver"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/sqlfrag"
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

	p    *pgxpool.Pool
	perr error
	once sync.Once
}

func (a *pgAdapter) Dialect() adapter.Dialect {
	return &a.dialect
}

func (a *pgAdapter) DriverName() string {
	return "postgres"
}

func (a *pgAdapter) Connector() driver.DriverContext {
	return loggingdriver.Wrap(stdlib.GetPoolConnector(a.p).Driver(), a.DriverName(), func(err error) int {
		var pqerr *pgconn.PgError
		if errors.As(err, &pqerr) {
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

var notRuntimeParams = map[string]struct{}{
	"host":                 {},
	"port":                 {},
	"database":             {},
	"user":                 {},
	"password":             {},
	"passfile":             {},
	"connect_timeout":      {},
	"sslmode":              {},
	"sslkey":               {},
	"sslcert":              {},
	"sslrootcert":          {},
	"sslpassword":          {},
	"sslsni":               {},
	"krbspn":               {},
	"krbsrvname":           {},
	"target_session_attrs": {},
	"service":              {},
	"servicefile":          {},

	//
	"_ro": {},
}

func (a *pgAdapter) Open(ctx context.Context, dsn *url.URL) (adapter.Adapter, error) {
	if a.DriverName() != dsn.Scheme {
		return nil, fmt.Errorf("invalid schema %s", dsn)
	}

	connParams := url.Values{}
	poolParams := url.Values{}

	for k, vv := range dsn.Query() {
		// only allow not runtime params as conn params
		if _, ok := notRuntimeParams[k]; ok {
			connParams[k] = vv
		}
		poolParams[k] = vv
	}

	a.once.Do(func() {
		if !poolParams.Has("pool_max_conns") {
			poolParams.Set("pool_max_conns", "10")
		}

		if !poolParams.Has("pool_min_conns") {
			poolParams.Set("pool_min_conns", "5")
		}

		if !poolParams.Has("pool_max_conn_lifetime") {
			poolParams.Set("pool_max_conn_lifetime", "30m")
		}

		if !poolParams.Has("pool_max_conn_idle_time") {
			poolParams.Set("pool_max_conn_idle_time", "30m")
		}

		if !poolParams.Has("pool_health_check_period") {
			poolParams.Set("pool_health_check_period", "30s")
		}

		dsn.RawQuery = poolParams.Encode()

		p, err := pgxpool.New(ctx, dsn.String())
		if err != nil {
			a.perr = err
			return
		}
		a.p = p
	})
	if a.perr != nil {
		return nil, a.perr
	}

	dbName := dbNameFromDSN(dsn)

	dsn.RawQuery = connParams.Encode()

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
	var e *pgconn.PgError
	if errors.As(err, &e) {
		if e.Code == "23505" {
			return true
		}
	}
	return false
}

func isErrorUnknownDatabase(err error) bool {
	var e *pgconn.PgError
	if errors.As(err, &e) {
		if e.Code == "3D000" {
			return true
		}
	}
	return false
}

func (a *pgAdapter) createDatabase(ctx context.Context, dbName string, dsn url.URL) error {
	dsn.Path = "/postgres"

	adaptor, err := a.Open(ctx, &dsn)
	if err != nil {
		return err
	}
	defer adaptor.Close()

	_, err = adaptor.Exec(context.Background(), sqlfrag.Pair("CREATE DATABASE ?;", sqlfrag.Const(dbName)))
	return err
}
