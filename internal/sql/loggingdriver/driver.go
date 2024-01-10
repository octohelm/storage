package loggingdriver

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"database/sql/driver"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"
)

type ErrorLevel func(error error) int

func Wrap(d driver.Driver, name string, errorLevel func(error error) int) driver.DriverContext {
	return &loggerConnector{
		driver: d,
		opt: &opt{
			name:       name,
			errorLevel: errorLevel,
		},
	}
}

type opt struct {
	name       string
	errorLevel ErrorLevel
}

func (o opt) ErrorLevel(err error) int {
	if o.errorLevel != nil {
		return o.errorLevel(err)
	}
	return 1
}

type loggerConnector struct {
	driver driver.Driver
	opt    *opt
	dsn    string
}

func (c *loggerConnector) OpenConnector(dsn string) (driver.Connector, error) {
	return &loggerConnector{
		driver: c.driver,
		opt:    c.opt,
		dsn:    dsn,
	}, nil
}

func (c *loggerConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.Open(c.dsn)
}

func (c *loggerConnector) Driver() driver.Driver {
	return c
}

func (c *loggerConnector) Open(dsn string) (driver.Conn, error) {
	conn, err := c.driver.Open(dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open connection")
	}
	return &loggerConn{Conn: conn, opt: c.opt}, nil
}

var _ interface {
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
} = (*loggerConn)(nil)

type loggerConn struct {
	driver.Conn
	opt *opt
}

func (c *loggerConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *loggerConn) Prepare(query string) (driver.Stmt, error) {
	panic(fmt.Errorf("don't use Prepare"))
}

func (c *loggerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	_, logger := logr.Start(ctx, "SQLQuery")
	cost := startTimer()

	defer func() {
		q := interpolateParams(query, args)

		l := logger.WithValues("driver", c.opt.name, "sql", q)
		if err != nil {
			if c.opt.ErrorLevel(err) > 0 {
				l.Error(errors.Wrapf(err, "query failed"))
			} else {
				l.Warn(errors.Wrapf(err, "query failed"))
			}
		} else {
			l.WithValues("cost", cost().String()).Debug("")
		}

		logger.End()
	}()

	rows, err = c.Conn.(driver.QueryerContext).QueryContext(context.Background(), replaceValueHolder(query), args)
	return
}

func (c *loggerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (result driver.Result, err error) {
	cost := startTimer()
	_, logger := logr.Start(ctx, "SQLExec")

	defer func() {
		q := interpolateParams(query, args)
		l := logger.WithValues("driver", c.opt.name, "sql", q)
		if err != nil {
			if c.opt.ErrorLevel(err) > 0 {
				l.Error(errors.Wrap(err, "exec failed"))
			} else {
				l.Warn(errors.Wrapf(err, "exec failed"))
			}
		} else {
			l.WithValues("cost", cost().String()).Debug("")
		}

		logger.End()
	}()

	result, err = c.Conn.(driver.ExecerContext).ExecContext(context.Background(), replaceValueHolder(query), args)
	return
}

func replaceValueHolder(query string) string {
	index := 0
	data := []byte(query)

	e := bytes.NewBufferString("")

	for i := range data {
		c := data[i]
		switch c {
		case '?':
			e.WriteByte('$')
			e.WriteString(strconv.FormatInt(int64(index+1), 10))
			index++
		default:
			e.WriteByte(c)
		}
	}

	return e.String()
}

func startTimer() func() time.Duration {
	startTime := time.Now()
	return func() time.Duration {
		return time.Since(startTime)
	}
}

func (c *loggerConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	logger := logr.FromContext(ctx)

	logger.Debug("=========== Beginning Transaction ===========")

	// don't pass ctx into real driver to avoid connect discount
	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to begin transaction"))
		return nil, err
	}

	return &loggingTx{tx: tx, logger: logger}, nil
}

type loggingTx struct {
	logger logr.Logger
	tx     driver.Tx
}

func (tx *loggingTx) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		tx.logger.Debug("failed to commit transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Committed Transaction ===========")
	return nil
}

func (tx *loggingTx) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		tx.logger.Debug("failed to rollback transaction: %s", err)
		return err
	}
	tx.logger.Debug("=========== Rollback Transaction ===========")
	return nil
}
