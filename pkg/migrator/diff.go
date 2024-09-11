package migrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type action struct {
	typ   actionType
	exprs []sqlfrag.Fragment
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

	migrate := func(typ actionType, name string, exprs ...sqlfrag.Fragment) {
		if len(exprs) > 0 {

			switch typ {
			case dropTableIndex, addTableIndex:
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
	for nextCol := range nextTable.Cols() {
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
						continue
					}
					migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
					continue
				}

				prevColType, _ := sqlfrag.All(context.Background(), dialect.DataType(currentCol.Def()))
				currentColType, _ := sqlfrag.All(context.Background(), dialect.DataType(nextCol.Def()))

				if !strings.EqualFold(prevColType, currentColType) {
					colChanges[nextCol.Name()] = modifyTableColumn
					migrate(modifyTableColumn, nextCol.Name(), dialect.ModifyColumn(nextCol, currentCol))
				}
				continue
			}

			colChanges[nextCol.Name()] = dropTableColumn
			migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
		}

		if nextCol.Def().DeprecatedActions == nil {
			migrate(addTableColumn, nextCol.Name(), dialect.AddColumn(nextCol))
		}
	}

	for col := range currentTable.Cols() {
		// only drop tmp col
		// when need to drop real data col, must declare deprecated for migrate
		if strings.HasPrefix(col.Name(), "__") && nextTable.F(col.Name()) == nil {
			// drop column
			colChanges[col.Name()] = dropTableColumn
			migrate(dropTableColumn, col.Name(), dialect.DropColumn(col))
		}
	}

	for key := range nextTable.Keys() {
		name := key.Name()
		if key.IsPrimary() {
			// pkey could not change
			continue
		}

		for col := range key.Cols() {
			if tpe, ok := colChanges[col.Name()]; ok && tpe == modifyTableColumn {
				// always re index when col type modified
				migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
				migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
			}
		}

		prevKey := currentTable.K(name)
		if prevKey == nil {
			migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
		} else {
			if !key.IsPrimary() {
				indexDef, _ := sqlfrag.All(context.Background(), sqlbuilder.ColumnCollect(key.Cols()))
				prevIndexDef, _ := sqlfrag.All(context.Background(), sqlbuilder.ColumnCollect(prevKey.Cols()))

				if !strings.EqualFold(indexDef, prevIndexDef) {
					migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
					migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
				}
			}
		}

	}

	for key := range currentTable.Keys() {
		colDropped := false

		for col := range key.Cols() {
			if tpe, ok := colChanges[col.Name()]; ok && tpe == dropTableColumn {
				colDropped = true
				break
			}
		}

		if colDropped {
			// always drop related index when col drop
			migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))

			continue
		}

		if nextTable.K(key.Name()) == nil {
			// drop index not exists
			migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
		}
	}

	return
}
