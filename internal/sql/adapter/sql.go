package adapter

import (
	"context"
	"database/sql"

	contextx "github.com/octohelm/x/context"
)

type SqlDo interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type sqlDoContext struct {
}

func ContextWithSqlDo(ctx context.Context, db SqlDo) context.Context {
	return contextx.WithValue(ctx, sqlDoContext{}, db)
}

func SqlDoFromContext(ctx context.Context) SqlDo {
	sqlDo, ok := ctx.Value(sqlDoContext{}).(SqlDo)
	if ok {
		return sqlDo
	}
	return nil
}
