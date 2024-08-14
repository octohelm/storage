package filtergen

import (
	"cmp"
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
	tablegenutil "github.com/octohelm/storage/devpkg/tablegen/util"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func init() {
	gengo.Register(&filterGen{})
}

type filterGen struct {
}

func (*filterGen) Name() string {
	return "filter"
}

func (*filterGen) New(ctx gengo.Context) gengo.Generator {
	return &filterGen{}
}

func (g *filterGen) GenerateType(c gengo.Context, srcNamed *types.Named) error {
	if typStruct, ok := srcNamed.Underlying().(*types.Struct); ok {
		tables := make(map[*types.Named]sqlbuilder.Table)

		for i := range typStruct.NumFields() {
			f := typStruct.Field(i)
			tag := reflect.StructTag(typStruct.Tag(i))

			if named, ok := f.Type().(*types.Named); ok {
				t, err := tablegenutil.ScanTable(c, named)
				if err != nil {
					return err
				}

				tables[named] = t

				if s, ok := tag.Lookup("select"); ok {
					g.generateSubFilter(c, tables, cmp.Or(tag.Get("as"), tag.Get("by")), named, s)
				} else {
					g.generateIndexedFilter(c, t, named, tag.Get("domain"))
				}
			}
		}
	}

	return gengo.ErrSkip
}

func (g *filterGen) generateSubFilter(c gengo.Context, tables map[*types.Named]sqlbuilder.Table, as string, fromModeType *types.Named, fromModelFieldName string) {
	values := strings.Split(as, ".")

	modelTypeName := values[0]
	modelTypeFieldName := values[1]

	var modelType *types.Named

	for named := range tables {
		if named.Obj().Name() == modelTypeName {
			modelType = named
		}
	}

	if modelType == nil {
		return
	}

	c.Render(gengo.Snippet{gengo.T: `
func @ModelTypeName'By@ModelFieldName'From@FromModelTypeName'(ctx @contextContext, patchers ...@querierpatcherTyped[@FromModelType]) @querierpatcherTyped[@ModelType] {
	return @querierpatcherWhere[@ModelType](
		@ModelType'T.@ModelFieldName.V(
			@querierpatcherInSelectIfExists(ctx, @FromModelType'T.@FromModelFieldName, patchers...),
		),
	)
}
`,
		"ModelTypeName":  gengo.ID(modelType.Obj().Name()),
		"ModelType":      gengo.ID(modelType.String()),
		"ModelFieldName": gengo.ID(modelTypeFieldName),

		"FromModelTypeName":  gengo.ID(fromModeType.Obj().Name()),
		"FromModelType":      gengo.ID(fromModeType.String()),
		"FromModelFieldName": gengo.ID(fromModelFieldName),

		"contextContext":                 gengo.ID("context.Context"),
		"querierpatcherTyped":            gengo.ID("github.com/octohelm/storage/pkg/dal/compose/querierpatcher.Typed"),
		"querierpatcherWhere":            gengo.ID("github.com/octohelm/storage/pkg/dal/compose/querierpatcher.Where"),
		"querierpatcherInSelectIfExists": gengo.ID("github.com/octohelm/storage/pkg/dal/compose/querierpatcher.InSelectIfExists"),
	})
}

func (g *filterGen) generateIndexedFilter(c gengo.Context, t sqlbuilder.Table, named *types.Named, domainName string) {
	indexedFields := make([]string, 0)

	cols := map[string]bool{}

	for k, _ := range t.Keys().RangeKey {
		for col, _ := range k.Columns().RangeCol {
			cols[col.FieldName()] = true
		}

	}

	for col, _ := range t.Cols().RangeCol {
		if cols[col.FieldName()] {
			indexedFields = append(indexedFields, col.FieldName())
		}
	}

	if domainName == "" {
		domainName = strings.TrimPrefix(t.TableName(), "t_")
	}

	for _, fieldName := range indexedFields {
		f := t.F(fieldName)
		fieldType := f.Def().Type

		fieldComment := fmt.Sprintf("按 %s 筛选", func() string {
			if comment := f.Def().Comment; comment != "" {
				return comment
			}
			if list := f.Def().Description; len(list) > 0 {
				return list[0]
			}
			return ""
		}())

		c.Render(gengo.Snippet{gengo.T: `
type @ModelTypeName'By@FieldName struct {
	@composeFrom[@Type] 

	@fieldComment
	@FieldName *@filterFilter[@FieldType] ` + "`" + `name:"@domainName~@fieldName,omitempty" in:"query"` + "`" + `
}


func (f *@ModelTypeName'By@FieldName) Apply(q @dalQuerier) @dalQuerier {
	return @composeApplyQuerierFromFilter(q, @Type'T.@FieldName, f.@FieldName)
}
`,
			"ModelTypeName": gengo.ID(named.Obj().Name()),
			"Type":          gengo.ID(named.Obj()),

			"FieldName":    gengo.ID(fieldName),
			"FieldType":    gengo.ID(fieldType.String()),
			"fieldComment": gengo.Comment(fieldComment),
			"domainName":   gengo.ID(camelcase.LowerKebabCase(domainName)),
			"fieldName":    gengo.ID(camelcase.LowerCamelCase(fieldName)),

			"dalQuerier":                    gengo.ID("github.com/octohelm/storage/pkg/dal.Querier"),
			"composeFrom":                   gengo.ID("github.com/octohelm/storage/pkg/dal/compose.From"),
			"composeApplyQuerierFromFilter": gengo.ID("github.com/octohelm/storage/pkg/dal/compose.ApplyQuerierFromFilter"),
			"filterFilter":                  gengo.ID("github.com/octohelm/storage/pkg/filter.Filter"),
		})
	}

}
