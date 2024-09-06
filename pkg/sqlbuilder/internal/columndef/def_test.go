package columndef

import (
	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/types"
	"reflect"
	"testing"
)

func TestColumnTypeFromTypeAndTag(t *testing.T) {
	cases := map[string]*ColumnDef{
		`,deprecated=f_target_env_id`: &ColumnDef{
			Type:              types.FromRType(reflect.TypeOf(1)),
			DeprecatedActions: &DeprecatedActions{RenameTo: "f_target_env_id"},
		},
		`,autoincrement`: &ColumnDef{
			Type:          types.FromRType(reflect.TypeOf(1)),
			AutoIncrement: true,
		},
		`,null`: &ColumnDef{
			Type: types.FromRType(reflect.TypeOf(float64(1.1))),
			Null: true,
		},
		`,size=2`: &ColumnDef{
			Type:   types.FromRType(reflect.TypeOf("")),
			Length: 2,
		},
		`,decimal=1`: &ColumnDef{
			Type:    types.FromRType(reflect.TypeOf(float64(1.1))),
			Decimal: 1,
		},
		`,default='1'`: &ColumnDef{
			Type:    types.FromRType(reflect.TypeOf("")),
			Default: ptr.String(`'1'`),
		},
	}

	for tagValue, ct := range cases {
		t.Run(tagValue, func(t *testing.T) {
			testingx.Expect(t, FromTypeAndTag(ct.Type, tagValue), testingx.Equal(ct))
		})
	}
}
