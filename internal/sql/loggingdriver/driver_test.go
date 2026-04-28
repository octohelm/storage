package loggingdriver

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
	. "github.com/octohelm/x/testing/v2"
)

type stubDriver struct {
	open func(name string) (driver.Conn, error)
}

func (d stubDriver) Open(name string) (driver.Conn, error) {
	return d.open(name)
}

type stubConn struct {
	query string
	exec  string
	tx    *stubTx
}

func (c *stubConn) Prepare(query string) (driver.Stmt, error) { return nil, nil }
func (c *stubConn) Close() error                              { return nil }
func (c *stubConn) Begin() (driver.Tx, error)                 { return c.tx, nil }
func (c *stubConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.tx == nil {
		c.tx = &stubTx{}
	}
	return c.tx, nil
}

func (c *stubConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.query = query
	return stubRows{}, nil
}

func (c *stubConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.exec = query
	return driver.RowsAffected(1), nil
}

type stubRows struct{}

func (stubRows) Columns() []string              { return []string{"f"} }
func (stubRows) Close() error                   { return nil }
func (stubRows) Next(dest []driver.Value) error { return io.EOF }

type stubTx struct {
	committed  bool
	rolledBack bool
	err        error
}

func (tx *stubTx) Commit() error {
	tx.committed = true
	return tx.err
}

func (tx *stubTx) Rollback() error {
	tx.rolledBack = true
	return tx.err
}

func TestWrapAndConnector(t *testing.T) {
	conn := Wrap(stubDriver{
		open: func(name string) (driver.Conn, error) {
			return &stubConn{}, nil
		},
	}, "primary", nil)

	roConnector, err := conn.OpenConnector("postgres://user:pass@localhost/db?_ro=true&sslmode=disable")
	Then(t, "OpenConnector 会移除 _ro 并在名字后追加 ::ro",
		Expect(err, Equal(error(nil))),
		Expect(roConnector.(*loggerConnector).opt.name, Equal("primary::ro")),
		Expect(roConnector.(*loggerConnector).dsn, Equal("postgres://user:pass@localhost/db?sslmode=disable")),
	)

	Then(t, "非法 dsn 返回错误",
		ExpectDo(func() error {
			_, err := conn.OpenConnector("://bad")
			return err
		}, ErrorMatch(regexp.MustCompile("missing protocol scheme"))),
	)
}

func TestLoggerConnAndHelpers(t *testing.T) {
	raw := &stubConn{tx: &stubTx{}}
	conn := &loggerConn{
		Conn: raw,
		opt:  &opt{name: "unit"},
	}

	_, err := conn.QueryContext(context.Background(), "SELECT ?", []driver.NamedValue{{Ordinal: 1, Value: 1}})
	Then(t, "QueryContext 会把 ? 改写为 $1",
		Expect(err, Equal(error(nil))),
		Expect(raw.query, Equal("SELECT $1")),
	)

	_, err = conn.ExecContext(context.Background(), "UPDATE t SET f = ?", []driver.NamedValue{{Ordinal: 1, Value: 2}})
	Then(t, "ExecContext 会把 ? 改写为 $1",
		Expect(err, Equal(error(nil))),
		Expect(raw.exec, Equal("UPDATE t SET f = $1")),
	)

	_, err = conn.BeginTx(context.Background(), driver.TxOptions{})
	Then(t, "BeginTx 返回 loggingTx 包装",
		Expect(err, Equal(error(nil))),
	)

	Then(t, "replaceValueHolder 与 startTimer 提供稳定辅助行为",
		Expect(replaceValueHolder("a ? b ?"), Equal("a $1 b $2")),
		ExpectMustValue(func() (bool, error) {
			timer := startTimer()
			time.Sleep(time.Millisecond)
			return timer() > 0, nil
		}, Equal(true)),
	)

	Then(t, "默认 ErrorLevel 为 1，显式函数可覆盖",
		Expect(opt{}.ErrorLevel(errors.New("x")), Equal(1)),
		Expect(opt{errorLevel: func(err error) int { return 3 }}.ErrorLevel(errors.New("x")), Equal(3)),
	)

	Then(t, "Prepare 明确禁止使用",
		ExpectMustValue(func() (panicked bool, err error) {
			defer func() {
				panicked = recover() != nil
			}()
			_, _ = conn.Prepare("SELECT 1")
			return panicked, nil
		}, Equal(true)),
	)
}

func TestLoggingTx(t *testing.T) {
	logger := logr.FromContext(logr.WithLogger(context.Background(), slog.Logger(slog.Default())))
	tx := &loggingTx{tx: &stubTx{}, logger: logger}
	Then(t, "Commit 与 Rollback 透传到底层事务",
		Expect(tx.Commit(), Equal(error(nil))),
		Expect(tx.Rollback(), Equal(error(nil))),
	)

	failTx := &loggingTx{tx: &stubTx{err: errors.New("tx failed")}, logger: logger}
	Then(t, "事务错误向上传递",
		ExpectDo(failTx.Commit, ErrorMatch(regexp.MustCompile("tx failed"))),
		ExpectDo(failTx.Rollback, ErrorMatch(regexp.MustCompile("tx failed"))),
	)
}

func mustRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
