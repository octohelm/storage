package columndef

import (
	"reflect"
	"testing"

	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/types"
)

func TestColumnTypeFromTypeAndTag(t *testing.T) {
	cases := map[string]*ColumnDef{
		`,deprecated=f_target_env_id`: {
			Type:              types.FromRType(reflect.TypeFor[int]()),
			DeprecatedActions: &DeprecatedActions{RenameTo: "f_target_env_id"},
		},
		`,autoincrement`: {
			Type:          types.FromRType(reflect.TypeFor[int]()),
			AutoIncrement: true,
		},
		`,null`: {
			Type: types.FromRType(reflect.TypeFor[float64]()),
			Null: true,
		},
		`,size=2`: {
			Type:   types.FromRType(reflect.TypeFor[string]()),
			Length: 2,
		},
		`,decimal=1`: {
			Type:    types.FromRType(reflect.TypeFor[float64]()),
			Decimal: 1,
		},
		`,default='1'`: {
			Type:    types.FromRType(reflect.TypeFor[string]()),
			Default: ptr.Ptr(`'1'`),
		},
	}

	for tagValue, ct := range cases {
		t.Run(tagValue, func(t *testing.T) {
			testingx.Expect(t, FromTypeAndTag(ct.Type, tagValue, ""), testingx.Equal(ct))
		})
	}
}
