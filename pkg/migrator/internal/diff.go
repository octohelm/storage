package internal

import (
	"cmp"
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
)

type actionType int

const (
	dropTableIndex actionType = iota
	dropTableColumn
	keepTableColumn
	renameTableColumn
	modifyTableColumn
	addTableColumn
	addTableIndex
	createTable
)

var _ sqlfrag.Fragment = &Action{}

type Action struct {
	typ       actionType
	name      string
	fragments []sqlfrag.Fragment
}

func (a *Action) IsNil() bool {
	return len(a.fragments) == 0
}

func (a *Action) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.JoinValues("", a.fragments...).Frag(ctx)
}

type diff struct {
	dialect adapter.Dialect
	actions []*Action

	indexes map[string]bool
	columns map[string]actionType
}

func (d *diff) IsNil() bool {
	return len(d.actions) == 0
}

func (d *diff) Frag(ctx context.Context) iter.Seq2[string, []any] {
	actions := slices.SortedFunc(slices.Values(d.actions), func(a *Action, b *Action) int {
		ret := cmp.Compare(a.typ, b.typ)
		if ret == 0 {
			return cmp.Compare(a.name, b.name)
		}
		return ret
	})

	return sqlfrag.Join("", sqlfrag.NonNil(slices.Values(actions))).Frag(ctx)
}

func (d *diff) migrate(typ actionType, name string, fragments ...sqlfrag.Fragment) {
	if len(fragments) > 0 {
		switch typ {
		case dropTableIndex, addTableIndex:
			// record once to avoid duplicated action
			changed := fmt.Sprintf("%d/%s", typ, name)
			if _, ok := d.indexes[changed]; ok {
				return
			} else {
				d.indexes[changed] = true
			}
		default:

		}

		d.actions = append(d.actions, &Action{
			typ:       typ,
			name:      name,
			fragments: fragments,
		})
	}
}

func Diff(dialect adapter.Dialect, currentTable sqlbuilder.Table, nextTable sqlbuilder.Table) sqlfrag.Fragment {
	d := &diff{
		dialect: dialect,
		indexes: make(map[string]bool),
		columns: make(map[string]actionType),
	}

	// create nextTable
	if currentTable == nil {
		d.migrate(createTable, nextTable.TableName(), dialect.CreateTableIsNotExists(nextTable)...)
		return d
	}

	// diff columns
	for nextCol := range nextTable.Cols() {
		if currentCol := currentTable.F(nextCol.Name()); currentCol != nil {
			if !nextCol.IsNil() {
				if deprecatedActions := sqlbuilder.GetColumnDef(nextCol).DeprecatedActions; deprecatedActions != nil {
					renameTo := deprecatedActions.RenameTo
					if renameTo != "" {
						if prevCol := currentTable.F(renameTo); prevCol != nil {
							d.columns[prevCol.Name()] = dropTableColumn
							d.migrate(dropTableColumn, prevCol.Name(), dialect.DropColumn(prevCol))
						}

						targetCol := nextTable.F(renameTo)
						if targetCol == nil {
							panic(fmt.Errorf("col `%s` is not declared", renameTo))
						}
						d.columns[renameTo] = renameTableColumn
						d.migrate(renameTableColumn, nextCol.Name(), dialect.RenameColumn(nextCol, targetCol))
						currentTable.(sqlbuilder.ColumnCollectionManger).AddCol(targetCol)
						continue
					}

					// drop deprecated column
					d.migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
					continue
				}

				prevColType, _ := sqlfrag.Collect(context.Background(), dialect.DataType(sqlbuilder.GetColumnDef(currentCol)))
				currentColType, _ := sqlfrag.Collect(context.Background(), dialect.DataType(sqlbuilder.GetColumnDef(nextCol)))

				if !strings.EqualFold(prevColType, currentColType) {
					d.columns[nextCol.Name()] = modifyTableColumn
					d.migrate(modifyTableColumn, nextCol.Name(), dialect.ModifyColumn(nextCol, currentCol))
				}

				d.columns[nextCol.Name()] = keepTableColumn

				continue
			}

			d.columns[nextCol.Name()] = dropTableColumn
			d.migrate(dropTableColumn, nextCol.Name(), dialect.DropColumn(nextCol))
		}
	}

	for col := range currentTable.Cols() {
		// only drop tmp col
		// when need to drop real data col, must declare deprecated for migrate
		if strings.HasPrefix(col.Name(), "__") && nextTable.F(col.Name()) == nil {
			// drop column
			d.columns[col.Name()] = dropTableColumn
			d.migrate(dropTableColumn, col.Name(), dialect.DropColumn(col))
		}
	}

	// added new columns
	for nextCol := range nextTable.Cols() {
		if sqlbuilder.GetColumnDef(nextCol).DeprecatedActions == nil {
			if _, changed := d.columns[nextCol.Name()]; !changed {
				d.migrate(addTableColumn, nextCol.Name(), dialect.AddColumn(nextCol))
			}
		}
	}

	for key := range nextTable.Keys() {
		name := key.Name()

		if key.IsPrimary() {
			if prevKey := currentTable.K(name); prevKey == nil {
				// prev key not exists, could create
				d.migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
			}
			// pkey could not change
			continue
		}

		for col := range key.Cols() {
			if tpe, ok := d.columns[col.Name()]; ok && tpe == modifyTableColumn {
				// always re index when col type modified
				d.migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
				d.migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
			}
		}

		prevKey := currentTable.K(name)
		if prevKey == nil {
			d.migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
		} else {
			if !key.IsPrimary() {
				currentIndexDef, _ := sqlfrag.Collect(context.Background(), sqlbuilder.ColumnCollect(
					slices.Values(slices.SortedFunc(key.Cols(), func(column1 sqlbuilder.Column, column2 sqlbuilder.Column) int {
						return cmp.Compare(column1.Name(), column2.Name())
					})),
				))
				prevIndexDef, _ := sqlfrag.Collect(context.Background(), sqlbuilder.ColumnCollect(slices.Values(slices.SortedFunc(prevKey.Cols(), func(column1 sqlbuilder.Column, column2 sqlbuilder.Column) int {
					return cmp.Compare(column1.Name(), column2.Name())
				}))))

				if !strings.EqualFold(currentIndexDef, prevIndexDef) {
					d.migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
					d.migrate(addTableIndex, key.Name(), dialect.AddIndex(key))
				}
			}
		}

	}

	for key := range currentTable.Keys() {
		colDropped := false

		for col := range key.Cols() {
			if tpe, ok := d.columns[col.Name()]; ok && tpe == dropTableColumn {
				colDropped = true
				break
			}
		}

		if colDropped {
			// always drop related index when col drop
			d.migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))

			continue
		}

		if nextTable.K(key.Name()) == nil {
			// drop index not exists
			if !key.IsPrimary() {
				d.migrate(dropTableIndex, key.Name(), dialect.DropIndex(key))
			}
		}
	}

	return d
}
