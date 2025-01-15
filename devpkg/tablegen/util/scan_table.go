package util

import (
	"context"
	"go/types"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	typesx "github.com/octohelm/x/types"
)

func toDefaultTableName(name string, tableGroup string) string {
	if tableGroup != "" && strings.ToLower(tableGroup) != strings.ToLower(name) {
		return gengo.LowerSnakeCase("t_" + tableGroup + "_" + name)
	}
	return gengo.LowerSnakeCase("t_" + name)
}

func ScanTable(c gengo.Context, named *types.Named) (sqlbuilder.Table, error) {
	tags, _ := c.Package(named.Obj().Pkg().Path()).Doc(named.Obj().Pos())

	tableGroup := ""

	if r, ok := tags["gengo:table:group"]; ok {
		if len(r) > 0 {
			tableGroup = r[0]
		}
	}

	tableName := toDefaultTableName(named.Obj().Name(), tableGroup)
	if tn, ok := tags["gengo:table:name"]; ok {
		if n := tn[0]; len(n) > 0 {
			tableName = n
		}
	}

	t := sqlbuilder.T(tableName)
	base := typesx.FromTType(named)

	getDoc := func(sf *typesx.TStructField) (map[string][]string, []string) {
		if pkg := sf.Pkg(); pkg != nil && pkg.Path() != "" {
			return c.Package(pkg.Path()).Doc(sf.Pos())
		}
		return c.Package(named.Obj().Pkg().Path()).Doc(sf.Pos())
	}

	for p := range structs.AllStructField(context.Background(), base) {
		def := p.ColumnType

		if tsf, ok := p.Field.(*typesx.TStructField); ok {
			tags, doc := getDoc(tsf)
			prefix := ""

			// embedded generics struct
			if len(p.Loc) > 1 {
				var ft typesx.Type = base

				for _, i := range p.Loc[0 : len(p.Loc)-1] {
					f := ft.Field(i).(*typesx.TStructField)

					if _, doc := getDoc(f); len(doc) > 0 {
						prefix += doc[0]
					}

					ft = f.Type()
					for ft.Kind() == reflect.Ptr {
						ft = ft.Elem()
					}
				}
			}

			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
			}

			def.Type = p.Type
			def.Comment, def.Description = commentAndDesc(doc)

			if values, ok := tags["rel"]; ok {
				rel := strings.Split(values[0], ".")
				if len(rel) >= 2 {
					def.Relation = rel
				}
			}
		}

		col := sqlbuilder.Col(p.Name, sqlbuilder.ColField(p.FieldName), sqlbuilder.ColDef(def))

		t.(sqlbuilder.ColumnCollectionManger).AddCol(col)
	}

	if indexes, ok := tags["def"]; ok {
		for i := range indexes {
			def := sqlbuilder.ParseIndexDefine(indexes[i])
			var key sqlbuilder.Key

			switch def.Kind {
			case "primary":
				key = sqlbuilder.PrimaryKey(nil, sqlbuilder.IndexFieldNameAndOptions(def.FieldNameAndOptions...))
			case "index":
				key = sqlbuilder.Index(def.Name, nil, sqlbuilder.IndexUsing(def.Method), sqlbuilder.IndexFieldNameAndOptions(def.FieldNameAndOptions...))
			case "unique_index":
				key = sqlbuilder.UniqueIndex(def.Name, nil, sqlbuilder.IndexUsing(def.Method), sqlbuilder.IndexFieldNameAndOptions(def.FieldNameAndOptions...))
			}

			if key != nil {
				t.(sqlbuilder.KeyCollectionManager).AddKey(key)
			}
		}
	}

	return t, nil
}

func commentAndDesc(docs []string) (comment string, desc []string) {
	for _, s := range docs {
		if comment == "" && s == "" {
			continue
		}
		if comment == "" {
			comment = s
		} else {
			desc = append(desc, s)
		}
	}
	return
}
