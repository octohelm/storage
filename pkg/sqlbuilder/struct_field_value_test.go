package sqlbuilder_test

import (
	"context"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/structs"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestFieldValueHelpersAndRename(t *testing.T) {
	user := &model.User{Name: "alice"}
	values := slices.Collect(structs.AllFieldValue(context.Background(), user))
	omitZero := slices.Collect(structs.AllFieldValueOmitZero(context.Background(), user, "Nickname"))

	Then(t, "AllFieldValue 和 OmitZero 暴露字段值",
		Expect(len(values) > 0, Equal(true)),
		Expect(len(omitZero) > 0, Equal(true)),
		Expect(omitZero[0].Field.FieldName != "", Equal(true)),
	)

	renamed := sqlbuilder.TableFromModel(&model.User{}).(sqlbuilder.TableWithTableName).WithTableName("t_user_shadow")
	colQ, _ := sqlfrag.Collect(sqlbuilder.ContextWithToggles(context.Background(), sqlbuilder.Toggles{
		sqlbuilder.ToggleMultiTable: true,
	}), renamed.F("Name"))
	Then(t, "WithTableName 会同步重绑定列和 key",
		Expect(renamed.TableName(), Equal("t_user_shadow")),
		Expect(renamed.K("primary") != nil, Equal(true)),
		Expect(colQ, Equal("t_user_shadow.f_name")),
	)
}
