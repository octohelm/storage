package sqlbuilder_test

import (
	"context"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestTablesAndKeys(t *testing.T) {
	tUser := sqlbuilder.T("t_user", sqlbuilder.Col("f_id", sqlbuilder.ColField("ID")), sqlbuilder.Col("f_name", sqlbuilder.ColField("Name")))
	tUser.(interface{ AddKey(...sqlbuilder.Key) }).AddKey(sqlbuilder.PrimaryKey(sqlbuilder.Cols("f_id")))

	k := tUser.K("primary")
	q, _ := sqlfrag.Collect(context.Background(), sqlbuilder.AsKeyColumnsTableDef(k))
	qOnly, _ := sqlfrag.Collect(context.Background(), sqlbuilder.AsKeyColumnsTableDef(k, sqlbuilder.KeyColumnOnly()))

	Then(
		t, "PrimaryKey 和 key 定义生成列列表",
		Expect(k.IsPrimary(), Equal(true)),
		Expect(k.IsUnique(), Equal(true)),
		Expect(k.Name(), Equal("primary")),
		Expect(q, Equal("f_id")),
		Expect(qOnly, Equal("f_id")),
		Expect(sqlbuilder.GetKeyTable(k).TableName(), Equal("t_user")),
		Expect(sqlbuilder.GetKeyDef(k).Method(), Equal("")),
	)

	custom := sqlbuilder.Index("i_name", sqlbuilder.Cols("f_name"), sqlbuilder.IndexUsing("btree"), sqlbuilder.IndexFieldNameAndOptions(sqlbuilder.FieldNameAndOption("f_name,desc")))
	custom = custom.Of(tUser)
	q, _ = sqlfrag.Collect(context.Background(), custom)
	Then(
		t, "自定义索引保留方法和列选项",
		Expect(q, Equal("i_name")),
		Expect(sqlbuilder.GetKeyDef(custom).Method(), Equal("btree")),
		Expect(len(slices.Collect(custom.Cols())), Equal(1)),
	)
}

func TestKeyCollectionAndCatalog(t *testing.T) {
	tUser := sqlbuilder.T("t_user", sqlbuilder.Col("f_id"))
	tUser.(interface{ AddKey(...sqlbuilder.Key) }).AddKey(
		sqlbuilder.PrimaryKey(sqlbuilder.Cols("f_id")),
		sqlbuilder.Index("i_name", sqlbuilder.Cols("f_id")),
	)
	Then(
		t, "KeyCollection 支持查找和遍历",
		Expect(tUser.K("PRIMARY").Name(), Equal("primary")),
		Expect(len(slices.Collect(tUser.Keys())), Equal(2)),
	)

	required := &sqlbuilder.Tables{}
	required.Add(sqlbuilder.TableFromModel(&model.Org{}))

	c := &sqlbuilder.Tables{}
	c.Add(tUser)
	c.Require(required)
	c.Remove("t_org")
	names := slices.Collect(sqlbuilder.TableNames(c))
	Then(
		t, "Catalog 支持增删查表和名字遍历",
		Expect(c.Table("t_user").TableName(), Equal("t_user")),
		Expect(c.Table("t_org"), Equal(sqlbuilder.Table(nil))),
		Expect(names, Equal([]string{"t_user", "t_org"})),
	)
}
