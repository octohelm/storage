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
	enumTypes EnumTypes
}

func (*enumGen) Name() string {
	return "enum"
}

func (*enumGen) New(ctx gengo.Context) gengo.Generator {
	g := &enumGen{
		enumTypes: EnumTypes{},
	}
	g.enumTypes.Walk(ctx, ctx.Package("").Pkg().Path())
	return g
}

func (g *enumGen) GenerateType(c gengo.Context, named *types.Named) error {
	if enum, ok := g.enumTypes.ResolveEnumType(named); ok {
		if enum.IsIntStringer() {
			g.genIntStringerEnums(c, named, enum)
			return nil
		}
		g.genEnums(c, named, enum)
	}
	return nil
}

func (g *enumGen) genEnums(c gengo.Context, named *types.Named, enum *EnumType) {
	options := make([]Option, len(enum.Constants))
	tpeObj := named.Obj()

	for i := range enum.Constants {
		options[i].Name = enum.Constants[i].Name()
		options[i].Value = fmt.Sprintf("%v", enum.Value(enum.Constants[i]))
		options[i].Label = enum.Label(enum.Constants[i])
	}

	c.Render(gengo.Snippet{
		gengo.T: `
var Invalid@Type = @errorsNew("invalid @Type")

func (@Type) EnumValues() []any {
	return []any{
		@constValues
	}
}
`,

		"Type":      gengo.ID(tpeObj),
		"errorsNew": gengo.ID("github.com/pkg/errors.New"),
		"constValues": gengo.MapSnippet(options, func(o Option) gengo.Snippet {
			return gengo.Snippet{
				gengo.T:     "@ConstName,",
				"ConstName": gengo.ID(o.Name),
			}
		}),
	})

	g.genLabel(c, tpeObj, enum, options)
}

type Option struct {
	Name  string
	Label string
	Value any
}

func (g *enumGen) genLabel(c gengo.Context, typ *types.TypeName, enum *EnumType, options []Option) {
	c.Render(gengo.Snippet{
		gengo.T: `
func Parse@Type'LabelString(label string) (@Type, error) {
	switch label {
		@labelToConstCases
		default:
			return @ConstUnknown, Invalid@Type
	}
}

func (v @Type) Label() string {
	switch v {
		@constToLabelCases
		default:
			return @fmtSprint(v)
	}
}

`,

		"Type": gengo.ID(typ.Name()),
		"ConstUnknown": func() gengo.Name {
			if enum.ConstUnknown != nil {
				return gengo.ID(enum.ConstUnknown)
			}
			return gengo.ID(`""`)
		}(),
		"fmtSprint": gengo.ID("fmt.Sprint"),
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
}

func (g *enumGen) genIntStringerEnums(c gengo.Context, tpe types.Type, enum *EnumType) {
	options := make([]Option, len(enum.Constants))
	tpeObj := tpe.(*types.Named).Obj()

	for i := range enum.Constants {
		options[i].Name = enum.Constants[i].Name()
		options[i].Value = fmt.Sprintf("%v", enum.Value(enum.Constants[i]))
		options[i].Label = enum.Label(enum.Constants[i])
	}

	c.Render(gengo.Snippet{
		gengo.T: `
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

	c.Render(gengo.Snippet{
		gengo.T: `
func (v @Type) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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
		default:
			var i @Type
			_, err := @fmtSscanf(s, "UNKNOWN_%d", &i)
			if err == nil {
				return i, nil
			}
			return @ConstUnknown, Invalid@Type
	}
}

func (v @Type) String() string {
	switch v {
		@constToStrValueCases
		case @ConstUnknown:
            return "UNKNOWN"
		default:
			return @fmtSprintf("UNKNOWN_%d", v)
	}
}

`,

		"Type": gengo.ID(tpeObj.Name()),
		"ConstUnknown": func() gengo.Name {
			if enum.ConstUnknown != nil {
				return gengo.ID(enum.ConstUnknown)
			}
			return gengo.ID(`""`)
		}(),
		"stringsHasPrefix": gengo.ID("strings.HasPrefix"),
		"fmtSscanf":        gengo.ID("fmt.Sscanf"),
		"fmtSprintf":       gengo.ID("fmt.Sprintf"),
		"bytesToUpper":     gengo.ID("bytes.ToUpper"),
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

	g.genLabel(c, tpeObj, enum, options)

	c.Render(gengo.Snippet{
		gengo.T: `
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
