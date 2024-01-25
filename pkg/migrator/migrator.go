package migrator

import (
	"context"
	"sort"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type actionType int

const (
	dropTableIndex actionType = iota
	dropTableColumn
	renameTableColumn
	modifyTableColumn
	addTableColumn
	addTableIndex
	createTable
)

func Migrate(ctx context.Context, a adapter.Adapter, toTables *sqlbuilder.Tables) error {
	fromTables, err := a.Catalog(ctx)
	if err != nil {
		return err
	}

	migrations := make([]sqlbuilder.SqlExpr, 0)

	for _, name := range toTables.TableNames() {
		as := diff(a.Dialect(), fromTables.Table(name), toTables.Table(name))
		sort.Sort(as)

		for i := range as {
			migrations = append(migrations, as[i].exprs...)
		}
	}

	if len(migrations) == 0 {
		return nil
	}

	return a.Transaction(ctx, func(ctx context.Context) error {
		for _, m := range migrations {
			if _, err := a.Exec(ctx, m); err != nil {
				return err
			}
		}
		return nil
	})
}
