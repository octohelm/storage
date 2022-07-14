package Table

import (
	"context"
	"go/types"
	"strings"

	"github.com/octohelm/storage/pkg/sqlbuilder"

	"github.com/octohelm/gengo/pkg/gengo"
	typesx "github.com/octohelm/x/types"
)

func init() {
	gengo.Register(&tableGen{})
}

type tableGen struct {
	gengo.SnippetWriter
}

func (tableGen) Name() string {
	return "table"
}

func (tableGen) New() gengo.Generator {
	return &tableGen{}
}

func (g *tableGen) Init(c *gengo.Context, s gengo.GeneratorCreator) (gengo.Generator, error) {
	return s.Init(c, g, func(g gengo.Generator, sw gengo.SnippetWriter) error {
		g.(*tableGen).SnippetWriter = sw
		return nil
	})
}

func toDefaultTableName(name string) string {
	return gengo.LowerSnakeCase("t_" + name)
}

func (g *tableGen) GenerateType(c *gengo.Context, named *types.Named) error {
	t := g.scanTable(c, named)

	g.generateIndexInterfaces(t, named)
	g.generateTableStatics(t, named)

	return nil
}

func (g *tableGen) generateConvertors(t sqlbuilder.Table, named *types.Named) {
	g.Do(`
func(v *[[ .typeName ]]) ColumnReceivers() map[string]interface{} {
	return map[string]interface{}{
		[[ .columnReceivers | render ]]
	}
}
`, gengo.Args{
		"typeName": named.Obj().Name(),
		"columnReceivers": func(sw gengo.SnippetWriter) {
			t.Cols().RangeCol(func(col sqlbuilder.Column, idx int) bool {
				sw.Do(`[[ .name | quote ]]: &v.[[ .fieldName ]],
`, gengo.Args{
					"name":      col.Name(),
					"fieldName": col.FieldName(),
				})
				return true
			})
		},
	})
}

func (g *tableGen) generateTableStatics(t sqlbuilder.Table, named *types.Named) {
	cols := t.Cols()
	keys := t.Keys()

	g.Do(`

type table[[ .typeName ]] struct {
	[[ .fieldNames | render ]]
	I indexNameOf[[ .typeName ]]
	table [[ "github.com/octohelm/storage/pkg/sqlbuilder.Table" | id ]]
}

func (table[[ .typeName ]]) New() [[ "github.com/octohelm/storage/pkg/sqlbuilder.Model" | id ]] {
	return &[[ .typeName ]]{}
}

func (t *table[[ .typeName ]]) IsNil() bool {
	return t.table.IsNil()
}

func (t *table[[ .typeName ]]) Ex(ctx [[ "context.Context" | id ]]) *[[ "github.com/octohelm/storage/pkg/sqlbuilder.Ex" | id ]]  {
	return t.table.Ex(ctx)
}

func (t *table[[ .typeName ]]) TableName() string {
	return t.table.TableName()
}

func (t *table[[ .typeName ]]) F(name string) [[ "github.com/octohelm/storage/pkg/sqlbuilder.Column" | id ]] {
	return t.table.F(name)
}

func (t *table[[ .typeName ]]) K(name string) [[ "github.com/octohelm/storage/pkg/sqlbuilder.Key" | id ]] {
	return t.table.K(name)
}

func (t *table[[ .typeName ]]) Cols(names ...string) [[ "github.com/octohelm/storage/pkg/sqlbuilder.ColumnCollection" | id ]] {
	return t.table.Cols(names...)
}

func (t *table[[ .typeName ]]) Keys(names ...string) [[ "github.com/octohelm/storage/pkg/sqlbuilder.KeyCollection" | id ]] {
	return t.table.Keys(names...)
}

type indexNameOf[[ .typeName ]] struct {
	[[ .indexNames | render ]]
}

var [[ .typeName ]]T = &table[[ .typeName ]]{
	[[ .fieldNameValues | render ]]
	I: indexNameOf[[ .typeName ]]{
		[[ .indexNameValues | render ]]
	},
	table: [[ "github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel" | id ]](&[[ .typeName ]]{}),
}
`, gengo.Args{
		"typeName": named.Obj().Name(),
		"fieldNames": func(sw gengo.SnippetWriter) {
			cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
				if col.Def().DeprecatedActions == nil {
					g.Do(`
[[ .fieldName ]] [[ "github.com/octohelm/storage/pkg/sqlbuilder.Column" | id ]]
`, gengo.Args{
						"fieldName": col.FieldName(),
					})
				}
				return true
			})
		},
		"fieldNameValues": func(sw gengo.SnippetWriter) {
			cols.RangeCol(func(col sqlbuilder.Column, idx int) bool {
				if col.Def().DeprecatedActions == nil {
					g.Do(`
[[ .fieldName ]]: [[ "github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel" | id ]](&[[ .typeName ]]{}).F([[ .fieldName | quote ]]),
`, gengo.Args{
						"typeName":  named.Obj().Name(),
						"fieldName": col.FieldName(),
					})
				}
				return true
			})
		},
		"indexNames": func(sw gengo.SnippetWriter) {
			keys.RangeKey(func(key sqlbuilder.Key, idx int) bool {
				if key.IsUnique() {
					g.Do(`
[[ .fieldName ]] [[ "github.com/octohelm/storage/pkg/sqlbuilder.ColumnCollection" | id ]]
`, gengo.Args{
						"fieldName": gengo.UpperCamelCase(key.Name()),
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

					g.Do(`
[[ .fieldName ]]: [[ "github.com/octohelm/storage/pkg/sqlbuilder.TableFromModel" | id ]](&[[ .typeName ]]{}).Cols([[ .names ]]...),
`, gengo.Args{
						"typeName":  named.Obj().Name(),
						"fieldName": gengo.UpperCamelCase(key.Name()),
						"names":     g.Dumper().ValueLit(names),
					})
				}
				return true
			})
		},
	})
}

func (g *tableGen) generateDescriptions(t sqlbuilder.Table, named *types.Named) {
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

	g.Do(`
[[ if .hasComments ]] func([[ .typeName ]]) Comments() map[string]string {
	return [[ .comments ]]
} [[ end ]]

[[ if .hasColDescriptions ]] func([[ .typeName ]]) ColDescriptions() map[string][]string {
	return [[ .colDescriptions ]]
} [[ end ]]

[[ if .hasColRelations ]] func([[ .typeName ]]) ColRelations() map[string][]string {
	return [[ .colRelations ]]
} [[ end ]]
`, gengo.Args{
		"typeName": named.Obj().Name(),

		"hasComments": len(colComments) > 0,
		"comments":    g.Dumper().ValueLit(colComments),

		"hasColDescriptions": len(colDescriptions) > 0,
		"colDescriptions":    g.Dumper().ValueLit(colDescriptions),

		"hasColRelations": len(colRelations) > 0,
		"colRelations":    g.Dumper().ValueLit(colRelations),
	})
}

func (g *tableGen) generateIndexInterfaces(t sqlbuilder.Table, named *types.Named) {
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

	g.Do(`
func ([[ .typeName ]]) TableName() string {
	return [[ .tableName | quote ]]
}

[[ if .hasPrimary ]] func([[ .typeName ]]) Primary() []string {
	return [[ .primary ]]
} [[ end ]]

[[ if .hasIndexes ]] func([[ .typeName ]]) Indexes() [[ "github.com/octohelm/storage/pkg/sqlbuilder.Indexes" | id ]] {
	return [[ .indexes ]]
} [[ end ]]

[[ if .hasUniqueIndexes ]] func([[ .typeName ]]) UniqueIndexes() [[ "github.com/octohelm/storage/pkg/sqlbuilder.Indexes" | id ]] {
	return [[ .uniqueIndexes ]]
} [[ end ]]
`, gengo.Args{
		"typeName":  named.Obj().Name(),
		"tableName": t.TableName(),

		"hasPrimary": len(primary) > 0,
		"primary":    g.Dumper().ValueLit(primary),

		"hasUniqueIndexes": len(uniqueIndexes) > 0,
		"uniqueIndexes":    g.Dumper().ValueLit(uniqueIndexes),

		"hasIndexes": len(indexes) > 0,
		"indexes":    g.Dumper().ValueLit(indexes),
	})
}

func (g *tableGen) scanTable(c *gengo.Context, named *types.Named) sqlbuilder.Table {
	tags, _ := c.Universe.Package(named.Obj().Pkg().Path()).Doc(named.Obj().Pos())

	tableName := toDefaultTableName(named.Obj().Name())
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
				tags, doc = c.Universe.Package(pkgPath).Doc(tsf.Pos())
			} else {
				tags, doc = c.Universe.Package(named.Obj().Pkg().Path()).Doc(tsf.Pos())
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

	return t
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
