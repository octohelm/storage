package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/octohelm/storage/pkg/sqlfrag"
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

func (d *db) Exec(ctx context.Context, frag sqlfrag.Fragment) (sql.Result, error) {
	if sqlfrag.IsNil(frag) {
		return nil, nil
	}

	query, args := sqlfrag.Collect(ctx, frag)
	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		result, err := sqlDo.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("exec failed: %w: %s", d.convertErr(err), query)
		}
		return result, nil
	}

	result, err := d.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w: %s", d.convertErr(err), query)
	}
	return result, nil
}

func (d *db) Query(ctx context.Context, frag sqlfrag.Fragment) (*sql.Rows, error) {
	if sqlfrag.IsNil(frag) {
		return nil, nil
	}
	query, args := sqlfrag.Collect(ctx, frag)

	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		rows, err := sqlDo.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w: %s", err, query)
		}
		return rows, err
	}

	rows, err := d.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w: %s", err, query)
	}
	return rows, err
}

func (d *db) Transaction(ctx context.Context, action func(ctx context.Context) error) (err error) {
	inScopeOfTxnCreated := false
	var txn *sql.Tx

	if sqlDo := SqlDoFromContext(ctx); sqlDo != nil {
		if tx, ok := sqlDo.(*sql.Tx); ok {
			txn = tx
		}
	}

	if txn == nil {
		tx, err := d.BeginTx(ctx, nil)
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
