package migrator

import (
	"context"
	"fmt"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator/internal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

func Migrate(ctx context.Context, a adapter.Adapter, toTables *sqlbuilder.Tables) error {
	fromTables, err := a.Catalog(ctx)
	if err != nil {
		return err
	}

	migrations := make([]sqlfrag.Fragment, 0)

	for _, name := range toTables.TableNames() {
		d := internal.Diff(a.Dialect(), fromTables.Table(name), toTables.Table(name))
		if sqlfrag.IsNil(d) {
			continue
		}

		migrations = append(migrations, d)
	}

	if len(migrations) == 0 {
		return nil
	}

	return a.Transaction(ctx, func(ctx context.Context) error {
		for _, m := range migrations {
			if _, err := a.Exec(ctx, m); err != nil {
				return fmt.Errorf("migrate failed: %w", err)
			}
		}
		return nil
	})
}
