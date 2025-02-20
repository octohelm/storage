package filtergen

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
	tablegenutil "github.com/octohelm/storage/devpkg/tablegen/util"
	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlpipe"
	sqlpipefilter "github.com/octohelm/storage/pkg/sqlpipe/filter"
)

func init() {
	gengo.Register(&filteropGen{})
}

type filteropGen struct{}

func (*filteropGen) Name() string {
	return "filterop"
}

func (*filteropGen) New(ctx gengo.Context) gengo.Generator {
	return &filteropGen{}
}

func (g *filteropGen) GenerateType(c gengo.Context, srcNamed *types.Named) error {
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

				if _, ok := tag.Lookup("select"); ok {
					// drop
				} else {
					g.generateIndexedFilter(c, t, named, tag.Get("domain"))
				}
			}
		}
	}

	return gengo.ErrSkip
}

func (g *filteropGen) generateIndexedFilter(c gengo.Context, t sqlbuilder.Table, named *types.Named, domainName string) {
	indexedFields := make([]string, 0)

	cols := map[string]bool{}

	for k := range t.Keys() {
		for col := range k.Cols() {
			cols[col.FieldName()] = true
		}
	}

	for col := range t.Cols() {
		if cols[col.FieldName()] {
			indexedFields = append(indexedFields, col.FieldName())
		}
	}

	modelTypeName := named.Obj().Name()

	domainFieldNameSuffix := ""
	if domainName != "" {
		domainType := camelcase.UpperCamelCase(domainName)
		if domainType != modelTypeName {
			domainFieldNameSuffix = "Of" + domainType
		}
	}

	if domainName == "" {
		domainName = strings.TrimPrefix(t.TableName(), "t_")
	}

	for _, fieldName := range indexedFields {
		f := t.F(fieldName)

		def := sqlbuilder.GetColumnDef(f)

		fieldType := def.Type

		fieldComment := fmt.Sprintf("%s", func() string {
			if comment := def.Comment; comment != "" {
				return fmt.Sprintf("通过 %s 筛选", comment)
			}
			return ""
		}())

		domainFieldName := camelcase.LowerCamelCase(fieldName)

		if jsonTag, ok := def.StructTag.Lookup("json"); ok {
			if jsonTag != "-" && jsonTag != "" {
				domainFieldName = strings.SplitN(jsonTag, ",", 2)[0]
			}
		}

		c.RenderT(`
type @ModelTypeName'By@DomainFieldName struct {
	@fieldComment
	@FieldName *@filterFilter[@FieldType] `+"`"+`name:"@domainName~@domainFieldName,omitzero" in:"query"`+"`"+`
}


func (f *@ModelTypeName'By@DomainFieldName) OperatorType() @sqlpipeOperatorType {
	return @sqlpipeOperatorFilter
}

func (f *@ModelTypeName'By@DomainFieldName) Next(src @sqlpipeSource[@Type]) @sqlpipeSource[@Type] {
	return src.Pipe(@sqlpipefilterAsWhere(@Type'T.@FieldName, f.@FieldName))
}
`, snippet.Args{
			"ModelTypeName": snippet.ID(modelTypeName),
			"Type":          snippet.ID(named.Obj()),

			"FieldType":       snippet.ID(fieldType.String()),
			"FieldName":       snippet.ID(fieldName),
			"fieldComment":    snippet.Comment(fieldComment),
			"domainName":      snippet.ID(camelcase.LowerKebabCase(domainName)),
			"DomainFieldName": snippet.ID(camelcase.UpperCamelCase(domainFieldName) + domainFieldNameSuffix),
			"domainFieldName": snippet.ID(domainFieldName),

			"sqlpipeSource":         snippet.PkgExposeFor[sqlpipe.P]("Source"),
			"sqlpipeOperatorType":   snippet.PkgExposeFor[sqlpipe.OperatorType](),
			"sqlpipeOperatorFilter": snippet.PkgExposeFor[sqlpipe.P]("OperatorFilter"),
			"sqlpipefilterAsWhere":  snippet.PkgExposeFor[sqlpipefilter.P]("AsWhere"),
			"filterFilter":          snippet.PkgExposeFor[filter.Filter[int]](),
		})

		_, sortable := def.StructTag.Lookup("sortable")
		if sortable {
			c.RenderT(`
type @ModelTypeName'SortBy@DomainFieldName struct {
}

func (f *@ModelTypeName'SortBy@DomainFieldName) Name() string {
	return "@domainName~@domainFieldName"
}

func (f *@ModelTypeName'SortBy@DomainFieldName) Label() string {
	return @sorterLabel
}

func (f *@ModelTypeName'SortBy@DomainFieldName) Sort(src @sqlpipeSource[@Type], sortBy func(col @sqlbuilderColumn) @sqlpipeSourceOperator[@Type]) @sqlpipeSource[@Type] {
	return src.Pipe(sortBy(@Type'T.@FieldName))
}
`, snippet.Args{
				"ModelTypeName": snippet.ID(modelTypeName),
				"Type":          snippet.ID(named.Obj()),

				"FieldType":       snippet.ID(fieldType.String()),
				"FieldName":       snippet.ID(fieldName),
				"domainName":      snippet.ID(camelcase.LowerKebabCase(domainName)),
				"DomainFieldName": snippet.ID(camelcase.UpperCamelCase(domainFieldName) + domainFieldNameSuffix),
				"domainFieldName": snippet.ID(domainFieldName),

				"sorterLabel": snippet.Value(def.Comment),

				"sqlpipeSource":         snippet.PkgExposeFor[sqlpipe.P]("Source"),
				"sqlpipeSourceOperator": snippet.PkgExposeFor[sqlpipe.P]("SourceOperator"),
				"sqlbuilderColumn":      snippet.PkgExposeFor[sqlbuilder.Column](),
				"sqlpipeOperatorType":   snippet.PkgExposeFor[sqlpipe.OperatorType](),
				"sqlpipeOperatorFilter": snippet.PkgExposeFor[sqlpipe.P]("OperatorFilter"),
				"sqlpipefilterAsWhere":  snippet.PkgExposeFor[sqlpipefilter.P]("AsWhere"),
				"filterFilter":          snippet.PkgExposeFor[filter.Filter[int]](),
			})
		}
	}
}
