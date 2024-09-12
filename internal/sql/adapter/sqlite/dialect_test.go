package sqlite

import (
	"context"
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"

	"github.com/octohelm/storage/pkg/sqlfrag"
	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestSqliteDialect(t *testing.T) {
	c := &dialect{}

	table := sqlbuilder.T("t",
		sqlbuilder.Col("f_id", sqlbuilder.ColTypeOf(uint64(0), ",autoincrement")),
		sqlbuilder.Col("f_old_name", sqlbuilder.ColTypeOf("", ",deprecated=f_name")),
		sqlbuilder.Col("f_name", sqlbuilder.ColTypeOf("", ",size=128,default=''")),
		sqlbuilder.Col("F_created_at", sqlbuilder.ColTypeOf(int64(0), ",default='0'")),
		sqlbuilder.Col("F_updated_at", sqlbuilder.ColTypeOf(int64(0), ",default='0'")),
		sqlbuilder.PrimaryKey(sqlbuilder.Cols("F_id")),
		sqlbuilder.UniqueIndex("I_name", sqlbuilder.Cols("F_id", "F_name"), sqlbuilder.IndexUsing("BTREE")),
		sqlbuilder.Index("I_created_at", sqlbuilder.Cols("F_created_at"), sqlbuilder.IndexUsing("BTREE")),
	)

	cases := map[string]struct {
		expr   sqlfrag.Fragment
		expect sqlfrag.Fragment
	}{
		"CreateTableIsNotExists": {
			c.CreateTableIsNotExists(table)[0],
			sqlfrag.Pair( /* language=sqlite */ `CREATE TABLE IF NOT EXISTS t (
	f_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	f_name TEXT NOT NULL DEFAULT '',
	f_created_at BIGINT NOT NULL DEFAULT '0',
	f_updated_at BIGINT NOT NULL DEFAULT '0'
);`),
		},
		"DropTable": {
			c.DropTable(table),
			sqlfrag.Pair( /* language=sqlite */ "DROP TABLE IF EXISTS t;"),
		},
		"TruncateTable": {
			c.TruncateTable(table),
			sqlfrag.Pair( /* language=sqlite */ "TRUNCATE TABLE t;"),
		},
		"AddColumn": {
			c.AddColumn(table.F("f_name")),
			sqlfrag.Pair( /* language=sqlite */ "ALTER TABLE t ADD COLUMN f_name TEXT NOT NULL DEFAULT '';"),
		},
		"DropColumn": {
			c.DropColumn(table.F("f_name")),
			sqlfrag.Pair( /* language=sqlite */ "ALTER TABLE t DROP COLUMN f_name;"),
		},
		"AddIndex": {
			c.AddIndex(table.K("I_name")),
			sqlfrag.Pair( /* language=sqlite */ "CREATE UNIQUE INDEX t_i_name ON t (f_id,f_name);"),
		},
		"AddPrimaryKey": {
			c.AddIndex(table.K("PRIMARY")),
			sqlfrag.Pair( /* language=sqlite */ "ALTER TABLE t ADD PRIMARY KEY (f_id);"),
		},
		"DropIndex": {
			c.DropIndex(table.K("I_name")),
			sqlfrag.Pair( /* language=sqlite */ "DROP INDEX IF EXISTS t_i_name;"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			q, args := sqlfrag.Collect(context.Background(), c.expect)

			testingx.Expect(t, c.expr, testutil.BeFragment(q, args...))
		})
	}
}
