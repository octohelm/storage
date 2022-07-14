package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/octohelm/storage/internal/sql/scanner"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type SqlExecutor = adapter.SqlDo

type DBExecutor interface {
	adapter.Adapter

	ExecExpr(expr sqlbuilder.SqlExpr) (sql.Result, error)
	QueryExpr(expr sqlbuilder.SqlExpr) (*sql.Rows, error)
	QueryExprAndScan(expr sqlbuilder.SqlExpr, v interface{}) error

	D() *Database
	T(model sqlbuilder.Model) sqlbuilder.Table

	Context() context.Context
	WithContext(ctx context.Context) DBExecutor
}

type db struct {
	ctx context.Context
	adapter.Adapter
	db *Database
}

func (d *db) WithContext(ctx context.Context) DBExecutor {
	return &db{
		ctx:     ctx,
		Adapter: d.Adapter,
		db:      d.db,
	}
}

func (d *db) Context() context.Context {
	if d.ctx != nil {
		return d.ctx
	}
	return context.Background()
}

func (d *db) D() *Database {
	return d.db
}

func (d *db) T(model sqlbuilder.Model) sqlbuilder.Table {
	return d.db.T(model)
}

func (d *db) ExecExpr(expr sqlbuilder.SqlExpr) (sql.Result, error) {
	return d.Exec(d.Context(), expr)
}

func (d *db) QueryExpr(expr sqlbuilder.SqlExpr) (*sql.Rows, error) {
	return d.Query(d.Context(), expr)
}

func (d *db) QueryExprAndScan(expr sqlbuilder.SqlExpr, v interface{}) error {
	ctx := d.Context()
	rows, err := d.Query(ctx, expr)
	if err != nil {
		return err
	}
	return scanner.Scan(ctx, rows, v)
}

func (d *db) SetMaxOpenConns(n int) {
	d.Adapter.(adapter.DBSetting).SetMaxOpenConns(n)
}

func (d *db) SetMaxIdleConns(n int) {
	d.Adapter.(adapter.DBSetting).SetMaxIdleConns(n)
}

func (d *db) SetConnMaxLifetime(t time.Duration) {
	d.Adapter.(adapter.DBSetting).SetConnMaxLifetime(t)
}
