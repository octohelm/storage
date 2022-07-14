package migrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type action struct {
	typ   actionType
	exprs []sqlbuilder.SqlExpr
}

type actions []action

func (a actions) Len() int {
	return len(a)
}

func (a actions) Less(i, j int) bool {
	return a[i].typ < a[j].typ
}

func (a actions) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func diff(dialect adapter.Dialect, currentTable sqlbuilder.Table, nextTable sqlbuilder.Table) (migrations actions) {
	indexes := map[string]bool{}

	migrate := func(typ actionType, name string, exprs ...sqlbuilder.SqlExpr) {
		if len(exprs) > 0 {

			switch typ {
			case addTableIndex:
				if _, ok := indexes[name]; ok {
					return
				} else {
					indexes[name] = true
				}
			}

			migrations = append(migrations, action{
				typ:   typ,
				exprs: exprs,
			})
		}
	}

	// create nextTable
	if currentTable == nil {
		migrate(createTable, nextTable.TableName(), dialect.CreateTableIsNotExists(nextTable)...)
		return
	}

	colChanges := map[string]actionType{}

	// diff columns
	nextTable.Cols().RangeCol(func(nextCol sqlbuilder.Column, idx int) bool {
		if currentCol := currentTable.F(nextCol.Name()); currentCol != nil {
			if nextCol != nil {
				if deprecatedActions := nextCol.Def().DeprecatedActions; deprecatedActions != nil {
					renameTo := deprecatedActions.RenameTo
					if renameTo != "" {
						prevCol := currentTable.F(renameTo)
						if prevCol != nil {
							colChanges[prevCol.Name()] = dropTableColumn
							migrate(dropTableColumn, prevCol.Name(), dialect.DropColumn(prevCol))
						}
						targetCol := nextTable.F(renameTo)
						if targetCol == nil {
							panic(fmt.Errorf("col `%s` is not declared", renameTo))
						}
						migrate(renameTableColumn, nextCol.Name(), dialect.RenameColumn(nextCol, targetCol))
						currentTable.(sqlbuilder.ColumnCollectionManger).AddCol(targetCol)
						return true
					}
					migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
					return true
				}

				prevColType := dialect.DataType(currentCol.Def()).Ex(context.Background()).Query()
				currentColType := dialect.DataType(nextCol.Def()).Ex(context.Background()).Query()

				if currentColType != prevColType {
					colChanges[nextCol.Name()] = modifyTableColumn
					migrate(modifyTableColumn, nextCol.Name(), dialect.ModifyColumn(nextCol, currentCol))
				}
				return true
			}

			colChanges[nextCol.Name()] = dropTableColumn
			migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
			return true
		}

		if nextCol.Def().DeprecatedActions == nil {
			migrate(addTableColumn, nextCol.Name(), dialect.AddColumn(nextCol))
		}

		return true
	})

	currentTable.Cols().RangeCol(func(col sqlbuilder.Column, idx int) bool {
		// only drop tmp col
		// when need to drop real data col, must declare deprecated for migrate
		if strings.HasPrefix(col.Name(), "__") && nextTable.F(col.Name()) == nil {
			// drop column
			colChanges[col.Name()] = dropTableColumn
			migrate(dropTableColumn, col.Name(), dialect.DropColumn(col))
		}
		return true
	})

	nextTable.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
		name := key.Name()
		if key.IsPrimary() {
			// pkey could not change
			return true
		}

		key.Columns().RangeCol(func(col sqlbuilder.Column, idx int) bool {
			if tpe, ok := colChanges[col.Name()]; ok && tpe == modifyTableColumn {
				// always re index when col type modified
				migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
				migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
			}
			return true
		})

		prevKey := currentTable.K(name)
		if prevKey == nil {
			migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
		} else {
			if !key.IsPrimary() {
				indexDef := key.Columns().Ex(context.Background()).Query()
				prevIndexDef := prevKey.Columns().Ex(context.Background()).Query()

				if !strings.EqualFold(indexDef, prevIndexDef) {
					migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
					migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
				}
			}
		}

		return true
	})

	currentTable.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
		colDropped := false

		key.Columns().RangeCol(func(col sqlbuilder.Column, idx int) bool {
			if tpe, ok := colChanges[col.Name()]; ok && tpe == dropTableColumn {
				colDropped = true
				return false
			}
			return true
		})

		if colDropped {
			// always drop related index when col drop
			migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
			return true
		}

		if nextTable.K(key.Name()) == nil {
			// drop index not exists
			migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
		}

		return true
	})

	return
}
