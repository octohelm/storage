package Table

import (
	"context"
	"go/types"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	typesx "github.com/octohelm/x/types"
)

func init() {
	gengo.Register(&tableGen{})
}

type tableGen struct {
}

func (*tableGen) Name() string {
	return "table"
}

func (*tableGen) New(ctx gengo.Context) gengo.Generator {
	return &tableGen{}
}

func toDefaultTableName(name string, tableGroup string) string {
	if tableGroup != "" && strings.ToLower(tableGroup) != strings.ToLower(name) {
		return gengo.LowerSnakeCase("t_" + tableGroup + "_" + name)
	}
	return gengo.LowerSnakeCase("t_" + name)
}

func (g *tableGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !named.Obj().Exported() {
		return gengo.ErrSkip
	}

	if _, ok := named.Underlying().(*types.Struct); ok {
		t, err := g.scanTable(c, named)
		if err != nil {
			return err
		}

		g.generateIndexInterfaces(c, t, named)
		g.generateTableStatics(c, t, named)

		return nil
	}

	return gengo.ErrSkip
}

func (g *tableGen) generateTableStatics(c gengo.Context, t sqlbuilder.Table, named *types.Named) {
	register := ""

	tags, _ := c.Doc(named.Obj())

	if r, ok := tags["gengo:table:register"]; ok {
		if len(r) > 0 {
			register = r[0]
		}
	}

	if register != "" {
		c.Render(gengo.Snippet{gengo.T: `
func init() {
	@Register.Add(@Type'T)
}

`,
			"Register": gengo.ID(register),
			"Type":     gengo.ID(named.Obj()),
		})
	}

	cols := t.Cols()
	keys := t.Keys()

	c.Render(gengo.Snippet{gengo.T: `
func (table@Type) New() @sqlbuilderModel {
	return &@Type{}
}

func (t *table@Type) IsNil() bool {
	return t.table.IsNil()
}

func (t *table@Type) Ex(ctx @contextContext) *@sqlbuilderEx  {
	return t.table.Ex(ctx)
}

func (t *table@Type) TableName() string {
	return t.table.TableName()
}

func (t *table@Type) F(name string) @sqlbuilderColumn {
	return t.table.F(name)
}

func (t *table@Type) K(name string) @sqlbuilderKey {
	return t.table.K(name)
}

func (t *table@Type) Cols(names ...string) @sqlbuilderColumnCollection {
	return t.table.Cols(names...)
}

func (t *table@Type) Keys(names ...string) @sqlbuilderKeyCollection {
	return t.table.Keys(names...)
}

type table@Type struct {
	I indexNameOf@Type
	table @sqlbuilderTable
	@FieldNames
}

type indexNameOf@Type struct {
	@indexNames
}

var @Type'T = &table@Type{
	@fieldNameValues
	I: indexNameOf@Type{
		@indexNameValues
	},
	table: @sqlbuilderTableFromModel(&@Type{}),
}
`,
		"Type": gengo.ID(named.Obj()),

		"contextContext":             gengo.ID("context.Context"),
		"sqlbuilderTableFromModel":   gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel"),
		"sqlbuilderEx":               gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Ex"),
		"sqlbuilderColumn":           gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Column"),
		"sqlbuilderColumnCollection": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.ColumnCollection"),
		"sqlbuilderKey":              gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Key"),
		"sqlbuilderKeyCollection":    gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.KeyCollection"),
		"sqlbuilderTable":            gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Table"),
		"sqlbuilderModel":            gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Model"),

		"FieldNames": func(sw gengo.SnippetWriter) {
			cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
				if col.Def().DeprecatedActions == nil {
					sw.Render(gengo.Snippet{gengo.T: `
@FieldName @sqlbuilderColumn
`,
						"FieldName":        gengo.ID(col.FieldName()),
						"sqlbuilderColumn": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Column"),
					})
				}
				return true
			})
		},

		"fieldNameValues": func(sw gengo.SnippetWriter) {
			cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
				if col.Def().DeprecatedActions == nil {
					sw.Render(gengo.Snippet{gengo.T: `
@FieldName: @sqlbuilderTableFromModel(&@Type{}).F(@FieldNameValue),
`,
						"Type":           gengo.ID(named.Obj()),
						"FieldName":      gengo.ID(col.FieldName()),
						"FieldNameValue": col.FieldName(),

						"sqlbuilderTableFromModel": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel"),
					})
				}
				return true
			})
		},

		"indexNames": func(sw gengo.SnippetWriter) {
			keys.RangeKey(func(key sqlbuilder.Key, idx int) bool {
				if key.IsUnique() {
					sw.Render(gengo.Snippet{gengo.T: `
@KeyName @sqlbuilderColumnCollection
`,
						"KeyName":                    gengo.ID(gengo.UpperCamelCase(key.Name())),
						"sqlbuilderColumnCollection": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.ColumnCollection"),
					})
				}
				return true
			})
		},

		"indexNameValues": func(sw gengo.SnippetWriter) {
			keys.RangeKey(func(key sqlbuilder.Key, idx int) bool {
				if key.IsUnique() {
					names := make([]string, 0)

					key.Columns().RangeCol(func(col sqlbuilder.Column, idx int) bool {
						names = append(names, col.FieldName())
						return true
					})

					sw.Render(gengo.Snippet{gengo.T: `
@KeyName: @sqlbuilderTableFromModel(&@Type{}).Cols(@keyNames...),
`,
						"KeyName":                  gengo.ID(gengo.UpperCamelCase(key.Name())),
						"Type":                     gengo.ID(named.Obj()),
						"sqlbuilderTableFromModel": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel"),
						"keyNames":                 names,
					})
				}
				return true
			})
		},
	})
}

func (g *tableGen) generateDescriptions(c gengo.Context, t sqlbuilder.Table, named *types.Named) {
	colComments := map[string]string{}
	colDescriptions := map[string][]string{}
	colRelations := map[string][]string{}

	t.Cols().RangeCol(func(col sqlbuilder.Column, idx int) bool {
		def := col.Def()

		if def.Comment != "" {
			colComments[col.FieldName()] = def.Comment
		}
		if len(def.Description) > 0 {
			colDescriptions[col.FieldName()] = col.Def().Description
		}
		if len(def.Relation) > 0 {
			colRelations[col.FieldName()] = def.Relation
		}

		return true
	})

	if len(colComments) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func(@Type) Comments() map[string]string {
	return @comments
}`,
			"Type":     gengo.ID(named.Obj()),
			"comments": colComments,
		})
	}

	if len(colDescriptions) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func(@Type) ColDescriptions() map[string][]string {
	return @colDescriptions
}`,
			"Type":            gengo.ID(named.Obj()),
			"colDescriptions": colDescriptions,
		})
	}

	if len(colRelations) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func(@Type) ColRelations() map[string][]string {
	return @colRelations
}`,
			"Type":         gengo.ID(named.Obj()),
			"colRelations": colRelations,
		})
	}
}

func (g *tableGen) generateIndexInterfaces(c gengo.Context, t sqlbuilder.Table, named *types.Named) {
	primary := make([]string, 0)
	indexes := sqlbuilder.Indexes{}
	uniqueIndexes := sqlbuilder.Indexes{}

	t.Keys().RangeKey(func(key sqlbuilder.Key, idx int) bool {
		keyDef := key.(sqlbuilder.KeyDef)

		if key.IsPrimary() {
			primary = keyDef.ColNameAndOptions()
		} else {
			n := key.Name()
			if method := keyDef.Method(); method != "" {
				n = n + "/" + method
			}
			if key.IsUnique() {
				uniqueIndexes[n] = keyDef.ColNameAndOptions()
			} else {
				indexes[n] = keyDef.ColNameAndOptions()
			}
		}
		return true
	})

	c.Render(gengo.Snippet{gengo.T: `
func (@Type) TableName() string {
	return @tableName
}

`,
		"Type":      gengo.ID(named.Obj()),
		"tableName": t.TableName(),
	})

	if len(primary) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func (@Type) Primary() []string {
	return @primary
}

`,
			"Type":    gengo.ID(named.Obj()),
			"primary": primary,
		})
	}

	if len(uniqueIndexes) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func (@Type) UniqueIndexes() @sqlbuilderIndexes {
	return @uniqueIndexes
}

`,
			"Type":              gengo.ID(named.Obj()),
			"sqlbuilderIndexes": gengo.ID(reflect.TypeOf(uniqueIndexes)),
			"uniqueIndexes":     uniqueIndexes,
		})
	}

	if len(indexes) > 0 {
		c.Render(gengo.Snippet{gengo.T: `
func (@Type) Indexes() @sqlbuilderIndexes {
	return @indexes
}

`,
			"Type":              gengo.ID(named.Obj()),
			"sqlbuilderIndexes": gengo.ID(reflect.TypeOf(indexes)),
			"indexes":           indexes,
		})
	}
}

func (g *tableGen) scanTable(c gengo.Context, named *types.Named) (sqlbuilder.Table, error) {
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
