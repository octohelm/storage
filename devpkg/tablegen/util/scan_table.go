package util

import (
	"context"
	"go/types"
	"strings"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/storage/pkg/sqlbuilder"
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

	sqlbuilder.EachStructField(context.Background(), typesx.FromTType(named), func(p *sqlbuilder.StructField) bool {
		def := sqlbuilder.ColumnDef{}

		if tsf, ok := p.Field.(*typesx.TStructField); ok {
			var tags map[string][]string
			var doc []string

			if pkgPath := p.Field.PkgPath(); pkgPath != "" {
				tags, doc = c.Package(pkgPath).Doc(tsf.Pos())
			} else {
				tags, doc = c.Package(named.Obj().Pkg().Path()).Doc(tsf.Pos())
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
		return true
	})

	if indexes, ok := tags["def"]; ok {
		for i := range indexes {
			def := sqlbuilder.ParseIndexDefine(indexes[i])
			var key sqlbuilder.Key

			switch def.Kind {
			case "primary":
				key = sqlbuilder.PrimaryKey(nil, sqlbuilder.IndexColNameAndOptions(def.ColNameAndOptions...))
			case "index":
				key = sqlbuilder.Index(def.Name, nil, sqlbuilder.IndexUsing(def.Method), sqlbuilder.IndexColNameAndOptions(def.ColNameAndOptions...))
			case "unique_index":
				key = sqlbuilder.UniqueIndex(def.Name, nil, sqlbuilder.IndexUsing(def.Method), sqlbuilder.IndexColNameAndOptions(def.ColNameAndOptions...))
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
