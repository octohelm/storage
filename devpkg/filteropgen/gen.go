package filtergen

import (
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

		c.Render(gengo.Snippet{
			gengo.T: `
type @ModelTypeName'By@FieldName struct {
	@fieldComment
	@FieldName *@filterFilter[@FieldType] ` + "`" + `name:"@domainName~@domainFieldName,omitzero" in:"query"` + "`" + `
}


func (f *@ModelTypeName'By@FieldName) OperatorType() @sqlpipeOperatorType {
	return @sqlpipeOperatorFilter
}

func (f *@ModelTypeName'By@FieldName) Next(src @sqlpipeSource[@Type]) @sqlpipeSource[@Type] {
	return src.Pipe(@sqlpipefilterAsWhere(@Type'T.@FieldName, f.@FieldName))
}
`,
			"ModelTypeName": gengo.ID(named.Obj().Name()),
			"Type":          gengo.ID(named.Obj()),

			"FieldName":       gengo.ID(fieldName),
			"FieldType":       gengo.ID(fieldType.String()),
			"fieldComment":    gengo.Comment(fieldComment),
			"domainName":      gengo.ID(camelcase.LowerKebabCase(domainName)),
			"domainFieldName": gengo.ID(domainFieldName),

			"sqlpipeSource":         gengo.ID("github.com/octohelm/storage/pkg/sqlpipe.Source"),
			"sqlpipeOperatorType":   gengo.ID("github.com/octohelm/storage/pkg/sqlpipe.OperatorType"),
			"sqlpipeOperatorFilter": gengo.ID("github.com/octohelm/storage/pkg/sqlpipe.OperatorFilter"),
			"sqlpipefilterAsWhere":  gengo.ID("github.com/octohelm/storage/pkg/sqlpipe/filter.AsWhere"),

			"filterFilter": gengo.ID("github.com/octohelm/storage/pkg/filter.Filter"),
		})
	}
}
