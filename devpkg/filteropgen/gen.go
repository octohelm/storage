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

	if domainName == "" {
		domainName = strings.TrimPrefix(t.TableName(), "t_")
	}

	for _, fieldName := range indexedFields {
		f := t.F(fieldName)

		def := sqlbuilder.GetColumnDef(f)
		fieldType := def.Type

		fieldComment := fmt.Sprintf("%s", func() string {
			if comment := def.Comment; comment != "" {
				return fmt.Sprintf("%s 通过 %s 筛选", fieldName, comment)
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
type @ModelTypeName'By@FieldName struct {
	@fieldComment
	@FieldName *@filterFilter[@FieldType] `+"`"+`name:"@domainName~@domainFieldName,omitzero" in:"query"`+"`"+`
}


func (f *@ModelTypeName'By@FieldName) OperatorType() @sqlpipeOperatorType {
	return @sqlpipeOperatorFilter
}

func (f *@ModelTypeName'By@FieldName) Next(src @sqlpipeSource[@Type]) @sqlpipeSource[@Type] {
	return src.Pipe(@sqlpipefilterAsWhere(@Type'T.@FieldName, f.@FieldName))
}
`, snippet.Args{
			"ModelTypeName": snippet.ID(named.Obj().Name()),
			"Type":          snippet.ID(named.Obj()),

			"FieldName":       snippet.ID(fieldName),
			"FieldType":       snippet.ID(fieldType.String()),
			"fieldComment":    snippet.Comment(fieldComment),
			"domainName":      snippet.ID(camelcase.LowerKebabCase(domainName)),
			"domainFieldName": snippet.ID(domainFieldName),

			"sqlpipeSource":         snippet.PkgExposeFor[sqlpipe.P]("Source"),
			"sqlpipeOperatorType":   snippet.PkgExposeFor[sqlpipe.OperatorType](),
			"sqlpipeOperatorFilter": snippet.PkgExposeFor[sqlpipe.P]("OperatorFilter"),
			"sqlpipefilterAsWhere":  snippet.PkgExposeFor[sqlpipefilter.P]("AsWhere"),
			"filterFilter":          snippet.PkgExposeFor[filter.Filter[int]](),
		})
	}
}
