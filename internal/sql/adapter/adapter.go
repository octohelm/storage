package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	syncx "github.com/octohelm/x/sync"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

// DB 定义了存储适配器依赖的最小数据库操作集合。
type DB interface {
	Exec(ctx context.Context, expr sqlfrag.Fragment) (sql.Result, error)
	Query(ctx context.Context, expr sqlfrag.Fragment) (*sql.Rows, error)
	Transaction(ctx context.Context, action func(ctx context.Context) error) error
	Close() error
}

// Connector 根据解析后的 DSN 打开一个 Adapter。
type Connector interface {
	Open(ctx context.Context, dsn *url.URL) (Adapter, error)
}

// Adapter 表示带有结构探查能力的驱动适配器。
type Adapter interface {
	DB
	DriverName() string
	Dialect() Dialect
	Catalog(ctx context.Context) (*sqlbuilder.Tables, error)
}

// Dialect 负责为具体数据库方言构造 DDL 片段。
type Dialect interface {
	CreateTableIsNotExists(t sqlbuilder.Table) []sqlfrag.Fragment
	DropTable(t sqlbuilder.Table) sqlfrag.Fragment
	TruncateTable(t sqlbuilder.Table) sqlfrag.Fragment

	AddColumn(col sqlbuilder.Column) sqlfrag.Fragment
	RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlfrag.Fragment
	ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlfrag.Fragment
	DropColumn(col sqlbuilder.Column) sqlfrag.Fragment

	AddIndex(key sqlbuilder.Key) sqlfrag.Fragment
	DropIndex(key sqlbuilder.Key) sqlfrag.Fragment

	DataType(columnDef sqlbuilder.ColumnDef) sqlfrag.Fragment
}

var adapters = syncx.Map[string, Adapter]{}

// Register 按驱动名及别名注册适配器。
func Register(a Adapter, aliases ...string) {
	adapters.Store(a.DriverName(), a)
	for i := range aliases {
		adapters.Store(aliases[i], a)
	}
}

// Open 根据 DSN scheme 选择已注册适配器并打开连接。
func Open(ctx context.Context, dsn string) (a Adapter, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	for key, value := range adapters.Range {
		if key == u.Scheme {
			a, err = value.(Connector).Open(ctx, u)
			break
		}
	}

	if a == nil && err == nil {
		return nil, fmt.Errorf("missing adapter for %s", u.Scheme)
	}

	return a, err
}
