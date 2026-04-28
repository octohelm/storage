package adapter

import (
	"context"
	"database/sql"

	contextx "github.com/octohelm/x/context"
)

// SqlDo 抽象了原始 SQL 辅助函数依赖的最小执行接口。
type SqlDo interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type sqlDoContext struct{}

// ContextWithSqlDo 向 context 注入 SqlDo，供下游辅助函数复用。
func ContextWithSqlDo(ctx context.Context, db SqlDo) context.Context {
	return contextx.WithValue(ctx, sqlDoContext{}, db)
}

// SqlDoFromContext 返回 context 中注入的 SqlDo。
func SqlDoFromContext(ctx context.Context) SqlDo {
	sqlDo, ok := ctx.Value(sqlDoContext{}).(SqlDo)
	if ok {
		return sqlDo
	}
	return nil
}
