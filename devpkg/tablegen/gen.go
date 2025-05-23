package tablegen

import (
	"go/types"
	"reflect"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	tablegenutil "github.com/octohelm/storage/devpkg/tablegen/util"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
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

	objectKind := named.Obj().Name()
	if r, ok := tags["gengo:table:objectkind"]; ok {
		if len(r) > 0 {
			objectKind = r[0]
		}
	}

	if register != "" {
		c.RenderT(`
func init() {
	@Register.Add(@Type'T)
}

`, snippet.Args{
			"Register": snippet.ID(register),
			"Type":     snippet.ID(named.Obj()),
		})
	}

	c.RenderT(`
func (@Type) GetKind() string {
	return @objectKind
}

`, snippet.Args{
		"Type":       snippet.ID(named.Obj()),
		"objectKind": snippet.Value(objectKind),
	})

	cols := t.Cols()

	c.RenderT(`
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
`, snippet.Args{
		"Type": snippet.ID(named.Obj()),

		"sqlbuilderModel": snippet.PkgExposeFor[sqlbuilder.P]("Model"),

		"modelScopedFromModel": snippet.PkgExposeFor[modelscoped.P]("FromModel"),
		"modelScopedTable":     snippet.PkgExposeFor[modelscoped.P]("Table"),
		"modelScopedKey":       snippet.PkgExposeFor[modelscoped.P]("Key"),

		"FieldNames": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for col := range cols {
				if def := sqlbuilder.GetColumnDef(col); def.DeprecatedActions == nil {
					if !yield(snippet.T(`
@fieldComment
@FieldName @modelScopedTypedColumn[@Type, @FieldType]
`, snippet.Args{
						"Type":                   snippet.ID(named.Obj()),
						"FieldName":              snippet.ID(col.FieldName()),
						"FieldType":              snippet.ID(def.Type.String()),
						"fieldComment":           snippet.Comment(def.Comment),
						"modelScopedTypedColumn": snippet.PkgExposeFor[modelscoped.P]("TypedColumn"),
					})) {
						return
					}
				}
			}
		}),

		"fieldNameValues": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for col := range cols {
				if def := sqlbuilder.GetColumnDef(col); def.DeprecatedActions == nil {
					if !yield(snippet.T(`
@FieldName: @modelScopedCastTypedColumn[@Type,@FieldType](@modelScopedFromModel[@Type]().F(@FieldNameValue)),
`, snippet.Args{
						"Type":           snippet.ID(named.Obj()),
						"FieldName":      snippet.ID(col.FieldName()),
						"FieldType":      snippet.ID(def.Type.String()),
						"FieldNameValue": snippet.Value(col.FieldName()),

						"modelScopedFromModel":       snippet.PkgExposeFor[modelscoped.P]("FromModel"),
						"modelScopedCastTypedColumn": snippet.PkgExposeFor[modelscoped.P]("CastTypedColumn"),
					})) {
						return
					}
				}
			}
		}),

		"indexNames": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for key := range t.Keys() {
				if key.IsUnique() {
					if !yield(snippet.T(`
@KeyName @modelScopedKey[@Type]
`, snippet.Args{
						"Type":    snippet.ID(named.Obj()),
						"KeyName": snippet.ID(gengo.UpperCamelCase(key.Name())),

						"modelScopedKey": snippet.PkgExposeFor[modelscoped.P]("Key"),
					})) {
						return
					}
				}
			}
		}),

		"indexValues": snippet.Snippets(func(yield func(snippet.Snippet) bool) {
			for key := range t.Keys() {
				if key.IsUnique() {
					if !yield(snippet.T(`
@KeyName: @modelScopedFromModel[@Type]().MK(@keyName),
`, snippet.Args{
						"KeyName": snippet.ID(gengo.UpperCamelCase(key.Name())),
						"keyName": snippet.Value(key.Name()),
						"Type":    snippet.ID(named.Obj()),

						"modelScopedFromModel": snippet.PkgExposeFor[modelscoped.P]("FromModel"),
					})) {
						return
					}
				}
			}
		}),
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
		c.RenderT(`
func(@Type) Comments() map[string]string {
	return @comments
}`, snippet.Args{
			"Type":     snippet.ID(named.Obj()),
			"comments": snippet.Value(colComments),
		})
	}

	if len(colDescriptions) > 0 {
		c.RenderT(`
func(@Type) ColDescriptions() map[string][]string {
	return @colDescriptions
}`, snippet.Args{
			"Type":            snippet.ID(named.Obj()),
			"colDescriptions": snippet.Value(colDescriptions),
		})
	}

	if len(colRelations) > 0 {
		c.RenderT(`
func(@Type) ColRelations() map[string][]string {
	return @colRelations
}`, snippet.Args{
			"Type":         snippet.ID(named.Obj()),
			"colRelations": snippet.Value(colRelations),
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

	c.RenderT(`
func (@Type) TableName() string {
	return @tableName
}
`, snippet.Args{
		"Type":      snippet.ID(named.Obj()),
		"tableName": snippet.Value(t.TableName()),
	})

	if len(primary) > 0 {
		c.RenderT(`
func (@Type) PrimaryKey() []string {
	return @primary
}

`, snippet.Args{
			"Type":    snippet.ID(named.Obj()),
			"primary": snippet.Value(primary),
		})
	}

	if len(uniqueIndexes) > 0 {
		c.RenderT(`
func (@Type) UniqueIndexes() @sqlbuilderIndexes {
	return @uniqueIndexes
}

`, snippet.Args{
			"Type":              snippet.ID(named.Obj()),
			"sqlbuilderIndexes": snippet.ID(reflect.TypeOf(uniqueIndexes)),
			"uniqueIndexes":     snippet.Value(uniqueIndexes),
		})
	}

	if len(indexes) > 0 {
		c.RenderT(`
func (@Type) Indexes() @sqlbuilderIndexes {
	return @indexes
}

`, snippet.Args{
			"Type":              snippet.ID(named.Obj()),
			"sqlbuilderIndexes": snippet.ID(reflect.TypeOf(indexes)),
			"indexes":           snippet.Value(indexes),
		})
	}
}
