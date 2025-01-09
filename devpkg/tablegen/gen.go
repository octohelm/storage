package tablegen

import (
	"go/types"
	"reflect"

	"github.com/octohelm/gengo/pkg/gengo"
	tablegenutil "github.com/octohelm/storage/devpkg/tablegen/util"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func init() {
	gengo.Register(&tableGen{})
}

type tableGen struct{}

func (*tableGen) Name() string {
	return "table"
}

func (*tableGen) New(ctx gengo.Context) gengo.Generator {
	return &tableGen{}
}

func (g *tableGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !named.Obj().Exported() {
		return gengo.ErrSkip
	}

	if _, ok := named.Underlying().(*types.Struct); ok {
		t, err := tablegenutil.ScanTable(c, named)
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
		c.Render(gengo.Snippet{
			gengo.T: `
func init() {
	@Register.Add(@Type'T)
}

`,
			"Register": gengo.ID(register),
			"Type":     gengo.ID(named.Obj()),
		})
	}

	cols := t.Cols()

	c.Render(gengo.Snippet{
		gengo.T: `
func (table@Type) New() @sqlbuilderModel {
	return &@Type{}
}

type table@Type struct {
	@modelScopedTable[@Type]

	I indexesOf@Type

	@FieldNames
}

type indexesOf@Type struct {
	@indexNames
}

var @Type'T = &table@Type{
	Table: @modelScopedFromModel[@Type](),

	@fieldNameValues

	I: indexesOf@Type{
		@indexValues
	},
}
`,
		"Type": gengo.ID(named.Obj()),

		"sqlbuilderModel": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder.Model"),

		"modelScopedFromModel": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.FromModel"),
		"modelScopedTable":     gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.Table"),
		"modelScopedKey":       gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.Key"),

		"FieldNames": func(sw gengo.SnippetWriter) {
			for col := range cols {
				if def := sqlbuilder.GetColumnDef(col); def.DeprecatedActions == nil {
					sw.Render(gengo.Snippet{
						gengo.T: `
@fieldComment
@FieldName @modelScopedTypedColumn[@Type, @FieldType]
`,
						"Type":         gengo.ID(named.Obj()),
						"FieldName":    gengo.ID(col.FieldName()),
						"FieldType":    gengo.ID(def.Type.String()),
						"fieldComment": gengo.Comment(def.Comment),

						"modelScopedTypedColumn": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.TypedColumn"),
					})
				}
			}
		},

		"fieldNameValues": func(sw gengo.SnippetWriter) {
			for col := range cols {
				if def := sqlbuilder.GetColumnDef(col); def.DeprecatedActions == nil {
					sw.Render(gengo.Snippet{
						gengo.T: `
@FieldName: @modelScopedCastTypedColumn[@Type,@FieldType](@modelScopedFromModel[@Type]().F(@FieldNameValue)),
`,
						"Type":           gengo.ID(named.Obj()),
						"FieldName":      gengo.ID(col.FieldName()),
						"FieldType":      gengo.ID(def.Type.String()),
						"FieldNameValue": col.FieldName(),

						"modelScopedFromModel":       gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.FromModel"),
						"modelScopedCastTypedColumn": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.CastTypedColumn"),
					})
				}
			}
		},

		"indexNames": func(sw gengo.SnippetWriter) {
			for key := range t.Keys() {
				if key.IsUnique() {
					sw.Render(gengo.Snippet{
						gengo.T: `
@KeyName @modelScopedKey[@Type]
`,
						"Type":    gengo.ID(named.Obj()),
						"KeyName": gengo.ID(gengo.UpperCamelCase(key.Name())),

						"modelScopedKey": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.Key"),
					})
				}
			}
		},

		"indexValues": func(sw gengo.SnippetWriter) {
			for key := range t.Keys() {
				if key.IsUnique() {
					sw.Render(gengo.Snippet{
						gengo.T: `
@KeyName: @modelScopedFromModel[@Type]().MK(@keyName),
`,
						"KeyName": gengo.ID(gengo.UpperCamelCase(key.Name())),
						"keyName": key.Name(),
						"Type":    gengo.ID(named.Obj()),

						"modelScopedFromModel": gengo.ID("github.com/octohelm/storage/pkg/sqlbuilder/modelscoped.FromModel"),
					})
				}
			}
		},
	})
}

func (g *tableGen) generateDescriptions(c gengo.Context, t sqlbuilder.Table, named *types.Named) {
	colComments := map[string]string{}
	colDescriptions := map[string][]string{}
	colRelations := map[string][]string{}

	for col := range t.Cols() {
		def := sqlbuilder.GetColumnDef(col)

		if def.Comment != "" {
			colComments[col.FieldName()] = def.Comment
		}
		if len(def.Description) > 0 {
			colDescriptions[col.FieldName()] = sqlbuilder.GetColumnDef(col).Description
		}
		if len(def.Relation) > 0 {
			colRelations[col.FieldName()] = def.Relation
		}
	}

	if len(colComments) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
func(@Type) Comments() map[string]string {
	return @comments
}`,
			"Type":     gengo.ID(named.Obj()),
			"comments": colComments,
		})
	}

	if len(colDescriptions) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
func(@Type) ColDescriptions() map[string][]string {
	return @colDescriptions
}`,
			"Type":            gengo.ID(named.Obj()),
			"colDescriptions": colDescriptions,
		})
	}

	if len(colRelations) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
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

	toStringSlice := func(options []sqlbuilder.FieldNameAndOption) []string {
		values := make([]string, 0, len(options))
		for _, o := range options {
			values = append(values, string(o))
		}
		return values
	}

	for key := range t.Keys() {
		keyDef := key.(sqlbuilder.KeyDef)
		fields := keyDef.FieldNameAndOptions()

		if key.IsPrimary() {
			primary = toStringSlice(fields)
		} else {
			n := key.Name()
			if method := keyDef.Method(); method != "" {
				n = n + "," + method
			}
			if key.IsUnique() {
				uniqueIndexes[n] = toStringSlice(fields)
			} else {
				indexes[n] = toStringSlice(fields)
			}
		}
	}

	c.Render(gengo.Snippet{
		gengo.T: `
func (@Type) TableName() string {
	return @tableName
}

`,
		"Type":      gengo.ID(named.Obj()),
		"tableName": t.TableName(),
	})

	if len(primary) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
func (@Type) PrimaryKey() []string {
	return @primary
}

`,
			"Type":    gengo.ID(named.Obj()),
			"primary": primary,
		})
	}

	if len(uniqueIndexes) > 0 {
		c.Render(gengo.Snippet{
			gengo.T: `
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
		c.Render(gengo.Snippet{
			gengo.T: `
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
