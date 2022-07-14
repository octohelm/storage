package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/octohelm/storage/pkg/sqlbuilder"

	"github.com/octohelm/storage/internal/testutil"
)

func TestPostgresDialect(t *testing.T) {
	c := &dialect{}

	table := sqlbuilder.T("t",
		sqlbuilder.Col("f_id", sqlbuilder.ColTypeOf(uint64(0), ",autoincrement")),
		sqlbuilder.Col("f_old_name", sqlbuilder.ColTypeOf("", ",deprecated=f_name")),
		sqlbuilder.Col("f_name", sqlbuilder.ColTypeOf("", ",size=128,default=''")),
		sqlbuilder.Col("f_geo", sqlbuilder.ColTypeOf(&Point{}, "")),
		sqlbuilder.Col("F_created_at", sqlbuilder.ColTypeOf(int64(0), ",default='0'")),
		sqlbuilder.Col("F_updated_at", sqlbuilder.ColTypeOf(int64(0), ",default='0'")),
		sqlbuilder.PrimaryKey(sqlbuilder.Cols("F_id")),
		sqlbuilder.UniqueIndex("I_name", sqlbuilder.Cols("F_id", "F_name"), sqlbuilder.IndexUsing("BTREE")),
		sqlbuilder.Index("I_created_at", sqlbuilder.Cols("F_created_at"), sqlbuilder.IndexUsing("BTREE")),
		sqlbuilder.Index("I_geo", sqlbuilder.Cols("F_geo"), sqlbuilder.IndexUsing("GIST")),
	)

	cases := map[string]struct {
		expr   sqlbuilder.SqlExpr
		expect sqlbuilder.SqlExpr
	}{
		"AddIndex": {
			c.AddIndex(table.K("I_name")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "CREATE UNIQUE INDEX t_i_name ON t USING BTREE (f_id,f_name);"),
		},
		"AddPrimaryKey": {
			c.AddIndex(table.K("PRIMARY")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "ALTER TABLE t ADD PRIMARY KEY (f_id);"),
		},
		"AddSpatialIndex": {
			c.AddIndex(table.K("i_geo")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "CREATE INDEX t_i_geo ON t USING GIST (f_geo);"),
		},
		"DropIndex": {
			c.DropIndex(table.K("i_name")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "DROP INDEX IF EXISTS t_i_name;"),
		},
		"DropPrimaryKey": {
			c.DropIndex(table.K("PRIMARY")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "ALTER TABLE t DROP CONSTRAINT t_pkey;"),
		},
		"CreateTableIsNotExists": {
			c.CreateTableIsNotExists(table)[0],
			sqlbuilder.Expr( /* language=PostgreSQL */ `CREATE TABLE IF NOT EXISTS t (
	f_id bigserial NOT NULL,
	f_name character varying(128) NOT NULL DEFAULT ''::character varying,
	f_geo POINT NOT NULL,
	f_created_at bigint NOT NULL DEFAULT '0'::bigint,
	f_updated_at bigint NOT NULL DEFAULT '0'::bigint,
	PRIMARY KEY (f_id)
);`),
		},
		"DropTable": {
			c.DropTable(table),
			sqlbuilder.Expr( /* language=PostgreSQL */ "DROP TABLE IF EXISTS t;"),
		},
		"TruncateTable": {
			c.TruncateTable(table),
			sqlbuilder.Expr( /* language=PostgreSQL */ "TRUNCATE TABLE t;"),
		},
		"AddColumn": {
			c.AddColumn(table.F("f_name")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "ALTER TABLE t ADD COLUMN f_name character varying(128) NOT NULL DEFAULT ''::character varying;"),
		},
		"DropColumn": {
			c.DropColumn(table.F("f_name")),
			sqlbuilder.Expr( /* language=PostgreSQL */ "ALTER TABLE t DROP COLUMN f_name;"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			testutil.ShouldBeExpr(t, c.expr, c.expect.Ex(context.Background()).Query())
		})
	}
}

type Point struct {
	X float64
	Y float64
}

func (Point) DataType(engine string) string {
	return "POINT"
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
