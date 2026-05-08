package adapter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

type fakeAdapter struct {
	DB
	dsn *url.URL
}

func (a *fakeAdapter) DriverName() string {
	return "unitadapter"
}

func (a *fakeAdapter) Dialect() Dialect {
	return nil
}

func (a *fakeAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	return nil, nil
}

func (a *fakeAdapter) Open(ctx context.Context, dsn *url.URL) (Adapter, error) {
	return &fakeAdapter{dsn: dsn}, nil
}

func TestRegisterAndOpen(t *testing.T) {
	Register(&fakeAdapter{}, "unitadapter-alias")

	opened, err := Open(context.Background(), "unitadapter-alias://host/db")
	Then(
		t, "Open 根据 scheme 找到已注册 adapter",
		Expect(err, Equal(error(nil))),
		Expect(opened.(*fakeAdapter).dsn.Scheme, Equal("unitadapter-alias")),
	)

	_, err = Open(context.Background(), "missing://host/db")
	Then(
		t, "Open 对未知 scheme 返回明确错误",
		Expect(err.Error(), Equal("missing adapter for missing")),
	)
}

func TestSqlDoContext(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := ContextWithSqlDo(context.Background(), db)

	Then(
		t, "ContextWithSqlDo 可写入并读取 SQL 执行器",
		Expect(SqlDoFromContext(ctx) == db, Equal(true)),
		Expect(SqlDoFromContext(context.Background()), Equal(SqlDo(nil))),
	)
}

func TestWrappedDBExecQueryAndTransaction(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	wrapped := Wrap(sqlDB, func(err error) error {
		return fmt.Errorf("converted: %w", err)
	})

	ctx := context.Background()
	mock.ExpectExec("UPDATE t SET f = \\?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	result, err := wrapped.Exec(ctx, sqlfrag.Pair("UPDATE t SET f = ?", 1))
	rowsAffected, _ := result.RowsAffected()

	Then(
		t, "Exec 收集 fragment 并执行 SQL",
		Expect(err, Equal(error(nil))),
		Expect(rowsAffected, Equal(int64(1))),
	)

	mock.ExpectQuery("SELECT f FROM t").WillReturnRows(sqlmock.NewRows([]string{"f"}).AddRow(1))
	rows, err := wrapped.Query(ctx, sqlfrag.Pair("SELECT f FROM t"))
	if rows != nil {
		defer rows.Close()
	}

	Then(
		t, "Query 收集 fragment 并返回 rows",
		Expect(err, Equal(error(nil))),
		Expect(rows != nil, Equal(true)),
	)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO t").WithArgs(2).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = wrapped.Transaction(ctx, func(ctx context.Context) error {
		_, err := wrapped.Exec(ctx, sqlfrag.Pair("INSERT INTO t VALUES (?)", 2))
		return err
	})

	Then(
		t, "Transaction 在无外层事务时提交成功动作",
		Expect(err, Equal(error(nil))),
	)

	mock.ExpectBegin()
	mock.ExpectRollback()
	errAction := fmt.Errorf("action failed")
	err = wrapped.Transaction(ctx, func(ctx context.Context) error {
		return errAction
	})

	Then(
		t, "Transaction 在动作返回错误时回滚",
		Expect(err, Equal(errAction)),
	)

	Then(
		t, "空 fragment 不执行 SQL",
		ExpectMustValue(func() (sql.Result, error) {
			return wrapped.Exec(ctx, nil)
		}, Equal(sql.Result(nil))),
		ExpectMustValue(func() (*sql.Rows, error) {
			return wrapped.Query(ctx, nil)
		}, Equal((*sql.Rows)(nil))),
	)

	Then(
		t, "sqlmock 所有预期均满足",
		ExpectDo(mock.ExpectationsWereMet),
	)
}

func TestWrappedDBErrors(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	wrapped := Wrap(sqlDB, func(err error) error {
		return fmt.Errorf("converted: %w", err)
	})

	mock.ExpectExec("DELETE FROM t").WillReturnError(driver.ErrBadConn)
	_, err = wrapped.Exec(context.Background(), sqlfrag.Pair("DELETE FROM t"))
	Then(
		t, "Exec 错误会经过 convertErr 包装",
		ExpectDo(func() error { return err }, ErrorMatch(regexp.MustCompile(`^exec failed: converted: .*: DELETE FROM t$`))),
	)

	mock.ExpectQuery("SELECT f FROM t").WillReturnError(driver.ErrBadConn)
	_, err = wrapped.Query(context.Background(), sqlfrag.Pair("SELECT f FROM t"))
	Then(
		t, "Query 错误保留原始错误",
		ExpectDo(func() error { return err }, ErrorMatch(regexp.MustCompile(`^query failed: .*: SELECT f FROM t$`))),
	)
}
