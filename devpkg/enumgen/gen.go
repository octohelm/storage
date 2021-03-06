package enumgen

import (
	"fmt"
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&enumGen{})
}

type enumGen struct {
	gengo.SnippetWriter
	enumTypes EnumTypes
}

func (*enumGen) Name() string {
	return "enum"
}

func (*enumGen) New(ctx gengo.Context) gengo.Generator {
	g := &enumGen{
		SnippetWriter: ctx.Writer(),
		enumTypes:     EnumTypes{},
	}

	g.enumTypes.Walk(ctx, ctx.Package("").Pkg().Path())

	return g
}

func (g *enumGen) GenerateType(c gengo.Context, named *types.Named) error {
	if enum, ok := g.enumTypes.ResolveEnumType(named); ok {
		if enum.IsIntStringer() {
			g.genIntStringerMethods(named, enum)
		}
	}
	return nil
}

type Option struct {
	Name  string
	Label string
	Value any
}

func (g *enumGen) genIntStringerMethods(tpe types.Type, enum *EnumType) {
	options := make([]Option, len(enum.Constants))

	tpeObj := tpe.(*types.Named).Obj()

	for i := range enum.Constants {
		options[i].Name = enum.Constants[i].Name()
		options[i].Value = fmt.Sprintf("%v", enum.Value(enum.Constants[i]))
		options[i].Label = enum.Label(enum.Constants[i])
	}

	g.Render(gengo.Snippet{gengo.T: `
var Invalid@Type = @errorsNew("invalid @Type")

func (@Type) EnumValues() []any {
	return []any{
		@constValues
	}
}
`,

		"Type":      gengo.ID(tpeObj.Name()),
		"errorsNew": gengo.ID("github.com/pkg/errors.New"),
		"constValues": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T:     "@ConstName,",
				"ConstName": gengo.ID(o.Name),
			}
		}),
	})

	g.Render(gengo.Snippet{gengo.T: `
func (v @Type) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, Invalid@Type
	}
	return []byte(str), nil
}

func (v *@Type) UnmarshalText(data []byte) (error) {
	vv, err := Parse@Type'FromString(string(@bytesToUpper(data)))
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

func Parse@Type'FromString(s string) (@Type, error) {
	switch s {
		@strValueToConstCases
	}
	return @ConstUnknown, Invalid@Type
}

func (v @Type) String() string {
	switch v {
		@constToStrValueCases
	}
	return "UNKNOWN"
}

`,

		"Type":         gengo.ID(tpeObj.Name()),
		"ConstUnknown": gengo.ID(enum.ConstUnknown),
		"bytesToUpper": gengo.ID("bytes.ToUpper"),
		"strValueToConstCases": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T: `
case @strValue:
	return @ConstName, nil
`,
				"strValue":  o.Value,
				"ConstName": gengo.ID(o.Name),
			}
		}),
		"constToStrValueCases": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T: `
case @ConstName:
	return @strValue
`,
				"strValue":  o.Value,
				"ConstName": gengo.ID(o.Name),
			}
		}),
	})

	g.Render(gengo.Snippet{gengo.T: `
func Parse@Type'LabelString(label string) (@Type, error) {
	switch label {
		@labelToConstCases
	}
	return @ConstUnknown, Invalid@Type
}

func (v @Type) Label() string {
	switch v {
		@constToLabelCases
	}
	return "UNKNOWN"
}

`,

		"Type":         gengo.ID(tpeObj.Name()),
		"ConstUnknown": gengo.ID(enum.ConstUnknown),
		"labelToConstCases": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T: `
case @labelValue:
	return @ConstName, nil
`,
				"labelValue": o.Label,
				"ConstName":  gengo.ID(o.Name),
			}
		}),
		"constToLabelCases": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T: `
case @ConstName:
	return @labelValue
`,
				"labelValue": o.Label,
				"ConstName":  gengo.ID(o.Name),
			}
		}),
	})

	g.Render(gengo.Snippet{gengo.T: `
func (v @Type) Value() (@driverValue, error) {
	offset := 0
	if o, ok := any(v).(@enumerationDriverValueOffset); ok {
		offset = o.Offset()
	}
	return int64(v) + int64(offset), nil
}

func (v *@Type) Scan(src any) error {
	offset := 0
	if o, ok := any(v).(@enumerationDriverValueOffset); ok {
		offset = o.Offset()
	}

	i, err := @enumerationScanIntEnumStringer(src, offset)
	if err != nil {
		return err
	}
	*v = @Type(i)
	return nil
}

`,
		"Type":                           gengo.ID(tpeObj),
		"driverValue":                    gengo.ID("database/sql/driver.Value"),
		"enumerationScanIntEnumStringer": gengo.ID("github.com/octohelm/storage/pkg/enumeration.ScanIntEnumStringer"),
		"enumerationDriverValueOffset":   gengo.ID("github.com/octohelm/storage/pkg/enumeration.DriverValueOffset"),
	})
}
