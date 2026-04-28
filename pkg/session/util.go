package session

import (
	"context"
	"database/sql"

	"github.com/octohelm/storage/internal/sql/adapter"
)

// InTx 判断当前 context 是否处于事务中。
func InTx(ctx context.Context) bool {
	sqlDo := adapter.SqlDoFromContext(ctx)
	if _, ok := sqlDo.(*sql.Tx); ok {
		return true
	}
	return false
}
