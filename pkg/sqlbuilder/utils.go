package sqlbuilder

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	typesx "github.com/octohelm/x/types"

	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
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
			fieldNameAndOptions: FieldNameAndOptionFromStringSlice(primaryKeyHook.PrimaryKey()),
		}).Of(tab))
	}

	if uniqueIndexesHook, ok := i.(WithUniqueIndexes); ok {
		for indexNameAndMethod, fieldNames := range uniqueIndexesHook.UniqueIndexes() {
			indexName, method := resolveIndexNameAndMethod(indexNameAndMethod)

			tab.KeyCollection.(KeyCollectionManager).AddKey((&key{
				name:                indexName,
				method:              method,
				isUnique:            true,
				fieldNameAndOptions: FieldNameAndOptionFromStringSlice(fieldNames),
			}).Of(tab))
		}
	}

	if indexesHook, ok := i.(WithIndexes); ok {
		for indexNameAndMethod, fieldNames := range indexesHook.Indexes() {
			indexName, method := resolveIndexNameAndMethod(indexNameAndMethod)
			tab.KeyCollection.(KeyCollectionManager).AddKey((&key{
				name:                indexName,
				method:              method,
				fieldNameAndOptions: FieldNameAndOptionFromStringSlice(fieldNames),
			}).Of(tab))
		}
	}
}

func resolveIndexNameAndMethod(n string) (name string, method string) {
	nameAndMethod := strings.Split(n, ",")
	name = strings.ToLower(nameAndMethod[0])
	if len(nameAndMethod) > 1 {
		method = nameAndMethod[1]
	}
	return name, method
}

// ParseIndexDefine
// @def index i_xxx,BTREE Name
// @def index i_xxx,GIST TEST,gist_trgm_ops
func ParseIndexDefine(def string) *IndexDefine {
	d := IndexDefine{}

	for i := strings.Index(def, " "); i != -1; i = strings.Index(def, " ") {
		part := def[0:i]

		if part != "" {
			if d.Kind == "" {
				d.Kind = part
			} else if d.Name == "" && d.Kind != "primary" {
				d.Name, d.Method = resolveIndexNameAndMethod(part)
			} else {
				break
			}
		}

		def = def[i+1:]
	}

	d.FieldNameAndOptions = FieldNameAndOptionFromStringSlice(strings.Split(strings.TrimSpace(def), " "))

	return &d
}

type IndexDefine struct {
	Kind                string
	Name                string
	Method              string
	FieldNameAndOptions []FieldNameAndOption
}

func (i IndexDefine) ID() string {
	if i.Method != "" {
		return i.Name + "," + i.Method
	}
	return i.Name
}

func FieldNameAndOptionFromStringSlice(slices []string) []FieldNameAndOption {
	fields := make([]FieldNameAndOption, 0, len(slices))

	for _, s := range slices {
		if s == "" {
			continue
		}
		fields = append(fields, FieldNameAndOption(s))
	}

	return fields
}

// {FieldName}[,DESC][,NULLS][,FIRST]
type FieldNameAndOption string

func (x FieldNameAndOption) Name() string {
	s := string(x)
	i := strings.Index(s, ",")
	if i > 0 {
		return pickFieldName(s[0:i])
	}
	return pickFieldName(s)
}

func pickFieldName(ref string) string {
	parts := strings.Split(ref, ".")
	return parts[len(parts)-1]
}

func (x FieldNameAndOption) Options() []string {
	s := string(x)
	i := strings.Index(s, ",")
	if i > 0 {
		return strings.Split(strings.ToUpper(s[i+1:]), ",")
	}
	return nil
}
