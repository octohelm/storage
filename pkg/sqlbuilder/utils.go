package sqlbuilder

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	typesx "github.com/octohelm/x/types"
)

func GetColumnName(fieldName, tagValue string) string {
	i := strings.Index(tagValue, ",")
	if tagValue != "" && (i > 0 || i == -1) {
		if i == -1 {
			return strings.ToLower(tagValue)
		}
		return strings.ToLower(tagValue[0:i])
	}
	return "f_" + strings.ToLower(fieldName)
}

var schemas sync.Map

func TableFromModel(model any) Table {
	tpe := reflect.TypeOf(model)
	if tpe.Kind() != reflect.Ptr {
		panic(fmt.Errorf("model %s must be a pointer", tpe.Name()))
	}
	tpe = tpe.Elem()
	if tpe.Kind() != reflect.Struct {
		panic(fmt.Errorf("model %s must be a struct", tpe.Name()))
	}

	if t, ok := schemas.Load(tpe); ok {
		if m, ok := model.(Model); ok {
			return t.(TableWithTableName).WithTableName(m.TableName())
		}
		return t.(Table)
	}

	tableName := tpe.Name()
	if m, ok := model.(Model); ok {
		tableName = m.TableName()
	}

	t := T(tableName).(*table)

	scanDefToTable(t, model)

	schemas.Store(tpe, t)

	return t
}

func scanDefToTable(tab *table, i any) {
	tpe := typesx.Deref(typesx.FromRType(reflect.TypeOf(i)))

	comments := map[string]string{}
	colDescriptions := map[string][]string{}
	colRelations := map[string][]string{}

	if withComments, ok := i.(WithComments); ok {
		comments = withComments.Comments()
	}

	if withColDescriptions, ok := i.(WithColDescriptions); ok {
		colDescriptions = withColDescriptions.ColDescriptions()
	}

	if withRelations, ok := i.(WithRelations); ok {
		colRelations = withRelations.ColRelations()
	}

	if tab.ColumnCollection == nil {
		tab.ColumnCollection = &columns{}
	}

	if tab.KeyCollection == nil {
		tab.KeyCollection = &keys{}
	}

	for f := range structs.AllStructField(context.Background(), tpe) {
		c := &column[any]{
			fieldName: f.FieldName,
			name:      f.Name,
			def:       f.ColumnType,
		}

		if comment, ok := comments[c.fieldName]; ok {
			c.def.Comment = comment
		}

		if desc, ok := colDescriptions[c.fieldName]; ok {
			c.def.Description = desc
		}

		if rel, ok := colRelations[c.fieldName]; ok {
			c.def.Relation = rel
		}

		tab.ColumnCollection.(ColumnCollectionManger).AddCol(c.Of(tab))
	}

	if withTableDescription, ok := i.(WithTableDescription); ok {
		desc := withTableDescription.TableDescription()
		tab.description = desc
	}

	if primaryKeyHook, ok := i.(WithPrimaryKey); ok {
		tab.KeyCollection.(KeyCollectionManager).AddKey((&key{
			name:                "primary",
			isUnique:            true,
			fieldNameAndOptions: primaryKeyHook.PrimaryKey(),
		}).Of(tab))
	}

	if uniqueIndexesHook, ok := i.(WithUniqueIndexes); ok {
		for indexNameAndMethod, fieldNames := range uniqueIndexesHook.UniqueIndexes() {
			indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)

			tab.KeyCollection.(KeyCollectionManager).AddKey((&key{
				name:                indexName,
				method:              method,
				isUnique:            true,
				fieldNameAndOptions: fieldNames,
			}).Of(tab))
		}
	}

	if indexesHook, ok := i.(WithIndexes); ok {
		for indexNameAndMethod, fieldNames := range indexesHook.Indexes() {
			indexName, method := ResolveIndexNameAndMethod(indexNameAndMethod)
			tab.KeyCollection.(KeyCollectionManager).AddKey((&key{
				name:                indexName,
				method:              method,
				fieldNameAndOptions: fieldNames,
			}).Of(tab))
		}
	}
}

func ResolveIndexNameAndMethod(n string) (name string, method string) {
	nameAndMethod := strings.Split(n, "/")
	name = strings.ToLower(nameAndMethod[0])
	if len(nameAndMethod) > 1 {
		method = nameAndMethod[1]
	}
	return
}

// ParseIndexDefine
// @def index i_xxx/BTREE Name
// @def index i_xxx/GIST TEST/gist_trgm_ops
func ParseIndexDefine(def string) *IndexDefine {
	d := IndexDefine{}

	for i := strings.Index(def, " "); i != -1; i = strings.Index(def, " ") {
		part := def[0:i]

		if part != "" {
			if d.Kind == "" {
				d.Kind = part
			} else if d.Name == "" && d.Kind != "primary" {
				d.Name, d.Method = ResolveIndexNameAndMethod(part)
			} else {
				break
			}
		}

		def = def[i+1:]
	}

	d.FieldNameAndOptions = strings.Split(strings.TrimSpace(def), " ")

	return &d
}

type IndexDefine struct {
	Kind                string
	Name                string
	Method              string
	FieldNameAndOptions []string
}

func (i IndexDefine) ID() string {
	if i.Method != "" {
		return i.Name + "/" + i.Method
	}
	return i.Name
}
