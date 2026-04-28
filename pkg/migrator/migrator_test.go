package migrator

import (
	"context"
	"database/sql"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	internaladapter "github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

type migratorDialect struct{}

func (migratorDialect) CreateTableIsNotExists(t sqlbuilder.Table) []sqlfrag.Fragment {
	return []sqlfrag.Fragment{sqlfrag.Const("CREATE TABLE " + t.TableName())}
}

func (migratorDialect) DropTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Const("DROP TABLE " + t.TableName())
}

func (migratorDialect) TruncateTable(t sqlbuilder.Table) sqlfrag.Fragment {
	return sqlfrag.Const("TRUNCATE " + t.TableName())
}

func (migratorDialect) AddColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Const("ADD COLUMN " + col.Name())
}

func (migratorDialect) RenameColumn(col sqlbuilder.Column, target sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Const("RENAME COLUMN " + col.Name() + " TO " + target.Name())
}

func (migratorDialect) ModifyColumn(col sqlbuilder.Column, prev sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Const("MODIFY COLUMN " + col.Name())
}

func (migratorDialect) DropColumn(col sqlbuilder.Column) sqlfrag.Fragment {
	return sqlfrag.Const("DROP COLUMN " + col.Name())
}

func (migratorDialect) AddIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	return sqlfrag.Const("ADD INDEX " + key.Name())
}

func (migratorDialect) DropIndex(key sqlbuilder.Key) sqlfrag.Fragment {
	return sqlfrag.Const("DROP INDEX " + key.Name())
}

func (migratorDialect) DataType(columnDef sqlbuilder.ColumnDef) sqlfrag.Fragment {
	return sqlfrag.Const(columnDef.DataType)
}

type migratorAdapter struct {
	catalog *sqlbuilder.Tables
	execed  int
}

func (a *migratorAdapter) Exec(ctx context.Context, expr sqlfrag.Fragment) (sql.Result, error) {
	a.execed++
	return nil, nil
}

func (a *migratorAdapter) Query(ctx context.Context, expr sqlfrag.Fragment) (*sql.Rows, error) {
	return nil, nil
}
func (a *migratorAdapter) Close() error                     { return nil }
func (a *migratorAdapter) DriverName() string               { return "sqlite" }
func (a *migratorAdapter) Dialect() internaladapter.Dialect { return migratorDialect{} }
func (a *migratorAdapter) Catalog(ctx context.Context) (*sqlbuilder.Tables, error) {
	return a.catalog, nil
}

func (a *migratorAdapter) Transaction(ctx context.Context, action func(ctx context.Context) error) error {
	return action(ctx)
}

func TestMigrateAndCreateTables(t *testing.T) {
	target := &sqlbuilder.Tables{}
	target.Add(sqlbuilder.TableFromModel(&model.User{}))

	a := &migratorAdapter{catalog: &sqlbuilder.Tables{}}
	Then(t, "Migrate 在目标表存在差异时执行迁移",
		ExpectDo(func() error {
			return Migrate(context.Background(), a, target)
		}),
	)
	Then(t, "Migrate 会执行迁移 SQL",
		Expect(a.execed > 0, Equal(true)),
	)

	a.execed = 0
	Then(t, "CreateTables 直接按目标 Catalog 创建",
		ExpectDo(func() error {
			return CreateTables(context.Background(), a, target)
		}),
	)
	Then(t, "CreateTables 会执行建表 SQL",
		Expect(a.execed > 0, Equal(true)),
	)
}
