// Package migrator 提供基于 catalog 差异的数据库结构迁移能力。
package migrator

import (
	"context"
	"fmt"
	"slices"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/migrator/internal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

// Migrate 按目标 catalog 执行数据库迁移。
func Migrate(ctx context.Context, a adapter.Adapter, toCatalog sqlbuilder.Catalog) error {
	fromTables, err := a.Catalog(ctx)
	if err != nil {
		return err
	}

	migrations := make([]sqlfrag.Fragment, 0)

	for _, name := range slices.Sorted(sqlbuilder.TableNames(toCatalog)) {
		d := internal.Diff(a.Dialect(), fromTables.Table(name), toCatalog.Table(name))
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

// CreateTables 仅按目标 catalog 创建缺失表结构。
func CreateTables(ctx context.Context, a adapter.Adapter, toCatalog sqlbuilder.Catalog) error {
	migrations := make([]sqlfrag.Fragment, 0)

	for _, name := range slices.Sorted(sqlbuilder.TableNames(toCatalog)) {
		d := internal.Diff(a.Dialect(), nil, toCatalog.Table(name))
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
