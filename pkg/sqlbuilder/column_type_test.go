package sqlbuilder_test

import (
	"reflect"
	"testing"

	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/x/ptr"
	"github.com/octohelm/x/types"
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
			testingx.Expect(t, ColumnDefFromTypeAndTag(ct.Type, tagValue), testingx.Equal(ct))
		})
	}
}
