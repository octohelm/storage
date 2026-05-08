package er

import (
	"regexp"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"

	. "github.com/octohelm/x/testing/v2"
)

func TestOrderedDatabaseJSON(t *testing.T) {
	raw := []byte(`{"tables":{"users":{"title":"Users","columns":{"id":{"type":"uint64"},"name":{"type":"string","of":"text"}},"constraints":{"primary":{"columnNames":[{"name":"id"}],"unique":true,"primary":true}}},"orgs":{"columns":{},"constraints":{}}}}`)

	var db OrderedDatabase
	Then(
		t, "OrderedDatabase 可从 JSON 解码",
		ExpectDo(func() error {
			return jsonv2.Unmarshal(raw, &db)
		}),
	)
	Then(
		t, "OrderedDatabase 按输入顺序解码表、列和约束",
		Expect(db.Tables.Len(), Equal(2)),
	)

	tables := make([]string, 0)
	for name := range db.Tables.KeyValues() {
		tables = append(tables, name)
	}

	users, ok := db.Tables.Get("users")
	Then(
		t, "record 支持有序遍历和按 key 获取",
		Expect(tables, Equal([]string{"users", "orgs"})),
		Expect(ok, Equal(true)),
		Expect(users.Title, Equal("Users")),
		Expect(users.Columns.Len(), Equal(2)),
	)

	users.Columns.Set("name", &OrderedColumn{Type: "varchar"})
	col, ok := users.Columns.Get("name")

	Then(
		t, "重复 Set 会替换值且不改变数量",
		Expect(users.Columns.Len(), Equal(2)),
		Expect(ok, Equal(true)),
		Expect(col.Type, Equal("varchar")),
		Expect(users.Columns.IsZero(), Equal(false)),
	)

	marshaled, err := jsonv2.Marshal(&db)
	Then(
		t, "MarshalJSONTo 保留 record 顺序",
		Expect(err, Equal(error(nil))),
		Expect(string(marshaled), Equal(`{"tables":{"users":{"title":"Users","columns":{"id":{"type":"uint64"},"name":{"type":"varchar"}},"constraints":{"primary":{"columnNames":[{"name":"id"}],"unique":true,"primary":true}}},"orgs":{"columns":{},"constraints":{}}}}`)),
	)

	Then(
		t, "兼容普通 ER 数据库结构",
		Expect(db.Er(), Equal(&db)),
		Expect(len(db.OneOf()), Equal(1)),
	)
}

func TestRecordJSONErrors(t *testing.T) {
	var r record[string, int]

	Then(
		t, "record 只接受对象 JSON",
		ExpectDo(func() error {
			return jsonv2.Unmarshal([]byte(`[]`), &r)
		}, ErrorMatch(mustRegexp("object should starts"))),
	)
}

func mustRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
