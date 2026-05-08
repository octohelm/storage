package modelscoped_test

import (
	"context"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/catalog"
	modelscoped "github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestModelScopedWrappers(t *testing.T) {
	tbl := modelscoped.FromModel[model.User]()
	cols := slices.Collect(tbl.MCols())
	keys := slices.Collect(tbl.MKeys())

	Then(
		t, "FromModel 暴露模型表、列和索引",
		Expect(tbl.TableName(), Equal("t_user")),
		Expect(len(cols) > 0, Equal(true)),
		Expect(len(keys) > 0, Equal(true)),
		Expect(tbl.MK("primary").Name(), Equal("primary")),
	)

	computed := model.UserT.Name.ComputedBy(model.UserT.Name)
	typedComputed := model.UserT.Name.TypedComputedBy(model.UserT.Name)
	Then(
		t, "列包装支持 ComputedBy",
		Expect(computed.FieldName(), Equal("Name")),
		Expect(typedComputed.FieldName(), Equal("Name")),
	)

	castTable := modelscoped.CastTable[model.User](tbl)
	castKey := modelscoped.CastKey[model.User](tbl.MK("primary"))
	castCol := modelscoped.CastColumn[model.User](tbl.F("Name"))
	castTypedCol := modelscoped.CastTypedColumn[model.User, string](tbl.F("Name"))
	qComputed, _ := sqlfrag.Collect(context.Background(), castTypedCol.TypedComputedBy(model.UserT.Name))
	Then(
		t, "包装类型暴露 Unwrap、Model 和 typed computed 行为",
		Expect(castTable.(interface{ Unwrap() sqlbuilder.Table }).Unwrap().TableName(), Equal("t_user")),
		Expect(castCol.(interface{ Unwrap() sqlbuilder.Column }).Unwrap().FieldName(), Equal("Name")),
		Expect(castTypedCol.(interface{ Unwrap() sqlbuilder.Column }).Unwrap().FieldName(), Equal("Name")),
		Expect(castTable.Model() != nil, Equal(true)),
		Expect(castKey.Model() != nil, Equal(true)),
		Expect(castCol.Model() != nil, Equal(true)),
		Expect(castTypedCol.Model() != nil, Equal(true)),
		Expect(qComputed, Equal("f_name")),
	)

	all := slices.Collect(modelscoped.AllColumns[model.User](model.UserT.Name, model.UserT.Age))
	Then(
		t, "AllColumns 保持顺序",
		Expect(len(all), Equal(2)),
		Expect(all[0].FieldName(), Equal("Name")),
	)

	c := catalog.From(&model.User{}, &model.Org{})
	Then(
		t, "catalog.From 把模型收集为 Catalog",
		Expect(c.Table("t_user").TableName(), Equal("t_user")),
		Expect(c.Table("t_org").TableName(), Equal("t_org")),
	)
}
