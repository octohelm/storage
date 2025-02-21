package sqlbuilder

import (
	"context"
	"iter"
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
