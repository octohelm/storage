package duckdb

import (
	"context"
	"testing"

	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
)

func TestDuckDBDialect(t *testing.T) {
	c := &dialect{}

	table := sqlbuilder.T("t",
		sqlbuilder.Col("f_id", sqlbuilder.ColTypeOf(uint64(0), ",autoincrement")),
		sqlbuilder.Col("f_old_name", sqlbuilder.ColTypeOf("", ",deprecated=f_name")),
		sqlbuilder.Col("f_name", sqlbuilder.ColTypeOf("", ",size=128,default=''")),
		sqlbuilder.Col("f_created_at", sqlbuilder.ColTypeOf(int64(0), ",default=0")),
		sqlbuilder.Col("f_updated_at", sqlbuilder.ColTypeOf(int64(0), ",default=0")),
		sqlbuilder.PrimaryKey(sqlbuilder.Cols("f_id")),
		sqlbuilder.UniqueIndex("I_name", sqlbuilder.Cols("f_id", "f_name"), sqlbuilder.IndexUsing("BTREE")),
		sqlbuilder.Index("i_created_at", sqlbuilder.Cols("f_created_at"), sqlbuilder.IndexUsing("BTREE")),
	)

	cases := map[string]struct {
		expr   sqlfrag.Fragment
		expect sqlfrag.Fragment
	}{
		"CreateTableIsNotExists": {
			c.CreateTableIsNotExists(table)[0],
			sqlfrag.Pair( /* language=duckdb */ `
CREATE SEQUENCE IF NOT EXISTS 'seq_t' START 1;
CREATE TABLE IF NOT EXISTS t (
	f_id INTEGER DEFAULT(nextval('seq_t')) PRIMARY KEY,
	f_name TEXT DEFAULT('') NOT NULL,
	f_created_at BIGINT DEFAULT(0) NOT NULL,
	f_updated_at BIGINT DEFAULT(0) NOT NULL
);
`),
		},
		"DropTable": {
			c.DropTable(table),
			sqlfrag.Pair( /* language=duckdb */ "DROP TABLE IF EXISTS t;"),
		},
		"AddColumn": {
			c.AddColumn(table.F("f_name")),
			sqlfrag.Pair( /* language=duckdb */ "ALTER TABLE t ADD COLUMN f_name TEXT DEFAULT('') NOT NULL;"),
		},
		"DropColumn": {
			c.DropColumn(table.F("f_name")),
			sqlfrag.Pair( /* language=duckdb */ "ALTER TABLE t DROP COLUMN f_name;"),
		},
		"AddIndex": {
			c.AddIndex(table.K("I_name")),
			sqlfrag.Pair( /* language=duckdb */ "CREATE UNIQUE INDEX t_i_name ON t (f_id,f_name);"),
		},
		"AddPrimaryKey": {
			c.AddIndex(table.K("PRIMARY")),
			sqlfrag.Pair( /* language=duckdb */ ""),
		},
		"DropIndex": {
			c.DropIndex(table.K("I_name")),
			sqlfrag.Pair( /* language=duckdb */ "DROP INDEX IF EXISTS t_i_name;"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			q, args := sqlfrag.Collect(context.Background(), c.expect)

			testingx.Expect(t, c.expr, testutil.BeFragment(q, args...))
		})
	}
}
