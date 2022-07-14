package enumgen

import (
	"fmt"
	"go/types"
	"strconv"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&enumGen{})
}

type enumGen struct {
	gengo.SnippetWriter
	enumTypes EnumTypes
}

func (enumGen) Name() string {
	return "enum"
}

func (enumGen) New() gengo.Generator {
	return &enumGen{
		enumTypes: EnumTypes{},
	}
}

func (g *enumGen) Init(c *gengo.Context, s gengo.GeneratorCreator) (gengo.Generator, error) {
	return s.Init(c, g, func(g gengo.Generator, sw gengo.SnippetWriter) error {
		g.(*enumGen).SnippetWriter = sw
		g.(*enumGen).enumTypes.Walk(c, c.Package.Pkg().Path())
		return nil
	})
}

func (g *enumGen) GenerateType(c *gengo.Context, named *types.Named) error {
	if enum, ok := g.enumTypes.ResolveEnumType(named); ok {
		if enum.IsIntStringer() {
			g.genIntStringerMethods(named, enum)
		}
	}
	return nil
}

func (g *enumGen) genIntStringerMethods(tpe types.Type, enum *EnumType) {
	options := make([]struct {
		Name        string
		QuotedValue string
		QuotedLabel string
	}, len(enum.Constants))

	tpeObj := tpe.(*types.Named).Obj()

	for i := range enum.Constants {
		options[i].Name = enum.Constants[i].Name()
		options[i].QuotedValue = strconv.Quote(fmt.Sprintf("%v", enum.Value(enum.Constants[i])))
		options[i].QuotedLabel = strconv.Quote(enum.Label(enum.Constants[i]))
	}

	a := gengo.Args{
		"typeName":    tpeObj.Name(),
		"typePkgPath": tpeObj.Pkg().Path(),

		"constUnknown": enum.ConstUnknown,
		"options":      options,
	}

	g.Do(`
var Invalid[[ .typeName ]] = [[ "github.com/pkg/errors.New" | id ]]("invalid [[ .typeName ]]")

func Parse[[ .typeName ]]FromString(s string) ([[ .typeName ]], error) {
	switch s {
	[[ range .options ]] case [[ .QuotedValue ]]:
		return [[ .Name ]], nil 
	[[ end ]] }
	return [[ .constUnknown | id ]], Invalid[[ .typeName ]]
}

func Parse[[ .typeName ]]FromLabelString(s string) ([[ .typeName ]], error) {
	switch s {
	[[ range .options ]] case [[ .QuotedLabel ]]:
		return [[ .Name ]], nil
	[[ end ]] }
	return [[ .constUnknown | id ]], Invalid[[ .typeName ]]
}

func ([[ .typeName ]]) TypeName() string {
	return "[[ .typePkgPath ]].[[ .typeName ]]"
}

func (v [[ .typeName ]]) String() string {
	switch v {
	[[ range .options ]] case [[ .Name ]]:
		return [[ .QuotedValue ]] 
	[[ end ]] }
	return "UNKNOWN"
}


func (v [[ .typeName ]]) Label() string {
	switch v {
	[[ range .options ]] case [[ .Name ]]:
		return [[ .QuotedLabel ]] 
	[[ end ]] }
	return "UNKNOWN"
}

func (v [[ .typeName ]]) Int() int {
	return int(v)
}

func ([[ .typeName ]]) ConstValues() [][[ "github.com/octohelm/storage/pkg/enumeration.IntStringerEnum" | id ]] {
	return [][[ "github.com/octohelm/storage/pkg/enumeration.IntStringerEnum" | id ]]{
		[[ range .options ]][[ .Name ]], 
		[[ end ]] }
}

func (v [[ .typeName ]]) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, Invalid[[ .typeName ]]
	}
	return []byte(str), nil
}

func (v *[[ .typeName ]]) UnmarshalText(data []byte) (err error) {
	*v, err = Parse[[ .typeName ]]FromString(string([[ "bytes.ToUpper" | id ]](data)))
	return
}


func (v [[ .typeName ]]) Value() ([[ "database/sql/driver.Value" | id ]], error) {
	offset := 0
	if o, ok := (interface{})(v).([[ "github.com/octohelm/storage/pkg/enumeration.DriverValueOffset" | id ]]); ok {
		offset = o.Offset()
	}
	return int64(v) + int64(offset), nil
}

func (v *[[ .typeName ]]) Scan(src interface{}) error {
	offset := 0
	if o, ok := (interface{})(v).([[ "github.com/octohelm/storage/pkg/enumeration.DriverValueOffset" | id ]]); ok {
		offset = o.Offset()
	}

	i, err := [[ "github.com/octohelm/storage/pkg/enumeration.ScanIntEnumStringer" | id ]](src, offset)
	if err != nil {
		return err
	}
	*v = [[ .typeName ]](i)
	return nil
}
`, a)
}
