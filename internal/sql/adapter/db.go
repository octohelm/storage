package adapter

import (
	"context"
	"database/sql"
	"runtime"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func Wrap(d *sql.DB, convertErr func(err error) error) DB {
	return &db{
		DB: d,
		option: option{
			convertErr: convertErr,
		},
	}
}

type option struct {
	convertErr func(err error) error
}

type db struct {
	option
	*sql.DB
}

func (d *db) Exec(ctx context.Context, expr sqlbuilder.SqlExpr) (sql.Result, error) {
	e := sqlbuilder.ResolveExprContext(ctx, expr)
	if sqlbuilder.IsNilExpr(e) {
		return nil, nil
	}
	if err := e.Err(); err != nil {
		return nil, d.convertErr(err)
	}

	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		result, err := sqlDo.ExecContext(ctx, e.Query(), e.Args()...)
		if err != nil {
			return nil, d.convertErr(err)
		}
		return result, nil
	}

	result, err := d.ExecContext(ctx, e.Query(), e.Args()...)
	if err != nil {
		return nil, d.convertErr(err)
	}
	return result, nil
}

func (d *db) Query(ctx context.Context, expr sqlbuilder.SqlExpr) (*sql.Rows, error) {
	e := sqlbuilder.ResolveExprContext(ctx, expr)
	if sqlbuilder.IsNilExpr(e) {
		return nil, nil
	}
	if err := e.Err(); err != nil {
		return nil, err
	}
	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		return sqlDo.QueryContext(ctx, e.Query(), e.Args()...)
	}
	return d.QueryContext(ctx, e.Query(), e.Args()...)
}

func (d *db) Transaction(ctx context.Context, action func(ctx context.Context) error) (err error) {
	var inScopeOfTxnCreated = false
	var txn *sql.Tx

	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		if tx, ok := sqlDo.(*sql.Tx); ok {
			txn = tx
		}
	}

	if txn == nil {
		tx, err := d.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		inScopeOfTxnCreated = true
		txn = tx
	}

	defer func() {
		if p := recover(); p != nil {
			// make sure rollack
			_ = txn.Rollback()

			switch e := p.(type) {
			case runtime.Error:
				panic(e)
			case error:
				err = e
			default:
				panic(e)
			}
		} else if inScopeOfTxnCreated {
			if err != nil {
				_ = txn.Rollback()
			} else {
				err = txn.Commit()
			}
		}
	}()

	return action(ContextWithSqlDo(ctx, txn))
}
