package sqlbuilder

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"sort"
	"text/scanner"
)

type TableDefinition interface {
	T() Table
}

type TableExprParse interface {
	Expr(query string, args ...any) *Ex
}

type TableWithTableName interface {
	WithTableName(name string) Table
}

type Table interface {
	SqlExpr

	// TableName of table
	TableName() string

	// K
	// get index by key name
	// primaryKey could be use `pk`
	K(k string) Key
	// F
	// get col by col name or struct field names
	F(name string) Column

	// Cols
	// get cols by col names or struct field name
	Cols(names ...string) ColumnCollection

	// Keys
	// get indexes by index names
	Keys(names ...string) KeyCollection
}

func T(tableName string, tableDefinitions ...TableDefinition) Table {
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

func (t *table) IsNil() bool {
	return t == nil || len(t.name) == 0
}

func (t *table) Ex(ctx context.Context) *Ex {
	return Expr(t.name).Ex(ctx)
}

func (t *table) Expr(query string, args ...any) *Ex {
	if query == "" {
		return nil
	}

	n := len(args)
	e := Expr("")
	e.Grow(n)

	s := &scanner.Scanner{}
	s.Init(bytes.NewBuffer([]byte(query)))

	queryCount := 0

	for tok := s.Next(); tok != scanner.EOF; tok = s.Next() {
		switch tok {
		case '#':
			fieldNameBuf := bytes.NewBuffer(nil)

			e.WriteHolder(0)

			for {
				tok = s.Next()

				if tok == scanner.EOF {
					break
				}

				if (tok >= 'A' && tok <= 'Z') ||
					(tok >= 'a' && tok <= 'z') ||
					(tok >= '0' && tok <= '9') ||
					tok == '_' {

					fieldNameBuf.WriteRune(tok)
					continue
				}

				e.WriteQueryByte(byte(tok))

				break
			}

			if fieldNameBuf.Len() == 0 {
				e.AppendArgs(t)
			} else {
				fieldName := fieldNameBuf.String()
				col := t.F(fieldNameBuf.String())
				if col == nil {
					panic(fmt.Errorf("missing field fieldName %s of table %s", fieldName, t.TableName()))
				}
				e.AppendArgs(col)
			}
		case '?':
			e.WriteQueryByte(byte(tok))
			if queryCount < n {
				e.AppendArgs(args[queryCount])
				queryCount++
			}
		default:
			e.WriteQueryByte(byte(tok))
		}
	}

	return e
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
