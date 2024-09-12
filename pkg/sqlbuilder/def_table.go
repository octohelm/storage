package sqlbuilder

import (
	"container/list"
	"context"
	"iter"
	"sort"
	"strings"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

type WithTable interface {
	T() Table
}

type TableCanFragment interface {
	Fragment(query string, args ...any) sqlfrag.Fragment

	// Deprecated
	// use Fragment instead
	Expr(query string, args ...any) sqlfrag.Fragment
}

type TableWithTableName interface {
	WithTableName(name string) Table
}

type Table interface {
	TableName() string

	KeyPicker
	KeySeq

	ColumnPicker
	ColumnSeq

	sqlfrag.Fragment
}

func T(tableName string, tableDefinitions ...sqlfrag.Fragment) Table {
	t := &table{
		name:             tableName,
		ColumnCollection: &columns{},
		KeyCollection:    &keys{},
	}

	// col added first
	for _, tableDef := range tableDefinitions {
		switch d := tableDef.(type) {
		case Column:
			t.AddCol(d.Of(t))
		}
	}

	for _, tableDef := range tableDefinitions {
		switch d := tableDef.(type) {
		case Key:
			t.AddKey(d.Of(t))
		}
	}

	return t
}

type table struct {
	database    string
	name        string
	description []string
	ColumnCollection
	KeyCollection
}

func (t *table) AddCol(cols ...Column) {
	for i := range cols {
		t.ColumnCollection.(ColumnCollectionManger).AddCol(cols[i].Of(t))
	}
}

func (t *table) AddKey(keys ...Key) {
	for i := range keys {
		t.KeyCollection.(KeyCollectionManager).AddKey(keys[i].Of(t))
	}
}

func (t table) WithTableName(name string) Table {
	newTable := &table{
		database:    t.database,
		name:        name,
		description: t.description,
	}

	newTable.ColumnCollection = t.ColumnCollection.Of(newTable)
	newTable.KeyCollection = t.KeyCollection.Of(newTable)

	return newTable
}

func (t *table) TableName() string {
	return t.name
}

func (t *table) String() string {
	return t.name
}

func (t *table) IsNil() bool {
	return t == nil || len(t.name) == 0
}

func (t *table) Frag(ctx context.Context) iter.Seq2[string, []any] {
	return sqlfrag.Pair(t.name).Frag(ctx)
}

func (t *table) Expr(query string, args ...any) sqlfrag.Fragment {
	return t.Fragment(query, args...)
}

func (t *table) Fragment(query string, args ...any) sqlfrag.Fragment {
	if query == "" {
		return nil
	}

	argSet := sqlfrag.NamedArgSet{
		"_t_": t,
	}

	for col := range t.ColumnCollection.Cols() {
		argSet["_t_"+col.FieldName()] = col
	}

	q := strings.ReplaceAll(query, "#", "@_t_")

	return sqlfrag.Pair(q, append([]any{argSet}, args...)...)
}

type Tables struct {
	l      *list.List
	tables map[string]*list.Element
}

func (tables *Tables) TableNames() (names []string) {
	tables.Range(func(tab Table, idx int) bool {
		names = append(names, tab.TableName())
		return true
	})
	sort.Strings(names)
	return
}

func (tables *Tables) Add(tabs ...Table) {
	if tables.tables == nil {
		tables.tables = map[string]*list.Element{}
		tables.l = list.New()
	}

	for _, tab := range tabs {
		if tab != nil {
			if _, ok := tables.tables[tab.TableName()]; ok {
				tables.Remove(tab.TableName())
			}
			e := tables.l.PushBack(tab)
			tables.tables[tab.TableName()] = e
		}
	}
}

func (tables *Tables) Table(tableName string) Table {
	if tables.tables != nil {
		if c, ok := tables.tables[tableName]; ok {
			return c.Value.(Table)
		}
	}
	return nil
}

func (tables *Tables) Remove(name string) {
	if tables.tables != nil {
		if e, exists := tables.tables[name]; exists {
			tables.l.Remove(e)
			delete(tables.tables, name)
		}
	}
}

func (tables Tables) Range(cb func(tab Table, idx int) bool) {
	if tables.l != nil {
		i := 0
		for e := tables.l.Front(); e != nil; e = e.Next() {
			if !cb(e.Value.(Table), i) {
				break
			}
			i++
		}
	}
}
