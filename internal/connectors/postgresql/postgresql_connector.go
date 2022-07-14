package postgresql

import (
	"context"
	"database/sql/driver"
	"io"

	"github.com/octohelm/sqlx/pkg/builder"
	"github.com/octohelm/sqlx/pkg/migration"
	"github.com/octohelm/sqlx/pkg/sqlx"

	"github.com/octohelm/sqlx/internal/connectors/postgresql/dialect"
)

var _ interface {
	driver.Connector
	builder.Dialect
} = (*PostgreSQLConnector)(nil)

type PostgreSQLConnector struct {
	dialect.Dialect
	Host       string
	DBName     string
	Extra      string
	Extensions []string
}

func (c *PostgreSQLConnector) Connect(ctx context.Context) (driver.Conn, error) {
	d := c.Driver()

	conn, err := d.Open(dsn(c.Host, c.DBName, c.Extra))
	if err != nil {
		if c.IsErrorUnknownDatabase(err) {
			connectForCreateDB, err := d.Open(dsn(c.Host, "", c.Extra))
			if err != nil {
				return nil, err
			}
			if _, err := connectForCreateDB.(driver.ExecerContext).ExecContext(context.Background(), builder.ResolveExpr(c.CreateDatabase(c.DBName)).Query(), nil); err != nil {
				return nil, err
			}
			if err := connectForCreateDB.Close(); err != nil {
				return nil, err
			}
			return c.Connect(ctx)
		}
		return nil, err
	}
	for _, ex := range c.Extensions {
		if _, err := conn.(driver.ExecerContext).ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS "+ex+";", nil); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (PostgreSQLConnector) Driver() driver.Driver {
	return &PostgreSQLLoggingDriver{}
}

func dsn(host string, dbName string, extra string) string {
	if extra != "" {
		extra = "?" + extra
	}
	return host + "/" + dbName + extra
}

func (c PostgreSQLConnector) WithDBName(dbName string) driver.Connector {
	c.DBName = dbName
	return &c
}

func (c *PostgreSQLConnector) Migrate(ctx context.Context, db sqlx.DBExecutor) error {
	output := migration.OutputFromContext(ctx)

	prevDB, err := dbFromInformationSchema(db)
	if err != nil {
		return err
	}

	d := db.D()
	dialect := db.Dialect()

	exec := func(expr builder.SqlExpr) error {
		if expr == nil || expr.IsNil() {
			return nil
		}

		if output != nil {
			_, _ = io.WriteString(output, builder.ResolveExpr(expr).Query())
			_, _ = io.WriteString(output, "\n")
			return nil
		}

		_, err := db.ExecExpr(expr)
		return err
	}

	if prevDB == nil {
		prevDB = &sqlx.Database{
			Name: d.Name,
		}
		if err := exec(dialect.CreateDatabase(d.Name)); err != nil {
			return err
		}
	}

	if d.Schema != "" {
		if err := exec(dialect.CreateSchema(d.Schema)); err != nil {
			return err
		}
		prevDB = prevDB.WithSchema(d.Schema)
	}

	for _, name := range d.Tables.TableNames() {
		table := d.Table(name)

		prevTable := prevDB.Table(name)

		if prevTable == nil {
			for _, expr := range dialect.CreateTableIsNotExists(table) {
				if err := exec(expr); err != nil {
					return err
				}
			}
			continue
		}

		exprList := table.Diff(prevTable, dialect)

		for _, expr := range exprList {
			if err := exec(expr); err != nil {
				return err
			}
		}
	}

	return nil
}
