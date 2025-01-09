package internal_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/sql/adapter/sqlite"
	"github.com/octohelm/storage/pkg/migrator/internal"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	"github.com/octohelm/storage/testdata/model"
	testingx "github.com/octohelm/x/testing"
)

func newAdapter(t testing.TB) adapter.Adapter {
	t.Helper()

	dir := t.TempDir()

	ctx := context.Background()

	u, _ := url.Parse(fmt.Sprintf("sqlite://%s", filepath.Join(dir, "sqlite.db")))

	a, err := sqlite.Open(ctx, u)
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		_ = a.Close()
		_ = os.RemoveAll(dir)
	})

	return a
}

func TestDiff(t *testing.T) {
	d := newAdapter(t)

	t.Run("init v1", func(t *testing.T) {
		userv1 := sqlbuilder.TableFromModel(&model.User{})
		actions := internal.Diff(d.Dialect(), nil, userv1)

		testingx.Expect(t, actions, testutil.BeFragment(`
CREATE TABLE IF NOT EXISTS t_user (
	f_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	f_name TEXT NOT NULL DEFAULT '',
	f_nickname TEXT NOT NULL DEFAULT '',
	f_username TEXT NOT NULL DEFAULT '',
	f_gender INTEGER NOT NULL DEFAULT '0',
	f_age BIGINT NOT NULL DEFAULT '0',
	f_created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	f_updated_at BIGINT NOT NULL DEFAULT '0',
	f_deleted_at BIGINT NOT NULL DEFAULT '0'
);
CREATE UNIQUE INDEX t_user_i_age ON t_user (f_age,f_deleted_at);
CREATE INDEX t_user_i_created_at ON t_user (f_created_at DESC);
CREATE UNIQUE INDEX t_user_i_name ON t_user (f_name,f_deleted_at);
CREATE INDEX t_user_i_nickname ON t_user (f_nickname);
`))
	})

	t.Run("init v2", func(t *testing.T) {
		userv2 := sqlbuilder.TableFromModel(&model.UserV2{})
		actions := internal.Diff(d.Dialect(), nil, userv2)

		testingx.Expect(t, actions, testutil.BeFragment(`
CREATE TABLE IF NOT EXISTS t_user (
	f_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	f_nickname TEXT NOT NULL DEFAULT '',
	f_gender INTEGER NOT NULL DEFAULT '0',
	f_real_name TEXT NOT NULL DEFAULT '',
	f_age INTEGER NOT NULL DEFAULT '0'
);
CREATE UNIQUE INDEX t_user_i_age ON t_user (f_age);
CREATE UNIQUE INDEX t_user_i_name ON t_user (f_real_name);
CREATE INDEX t_user_i_nickname ON t_user (f_nickname);
`))
	})

	t.Run("migrate from v1 to v2", func(t *testing.T) {
		userv1 := sqlbuilder.TableFromModel(&model.User{})
		userv2 := sqlbuilder.TableFromModel(&model.UserV2{})

		actions := internal.Diff(d.Dialect(), userv1, userv2)

		testingx.Expect(t, actions, testutil.BeFragment(`
DROP INDEX IF EXISTS t_user_i_age;
DROP INDEX IF EXISTS t_user_i_created_at;
DROP INDEX IF EXISTS t_user_i_name;
ALTER TABLE t_user DROP COLUMN f_username;
ALTER TABLE t_user RENAME COLUMN f_name TO f_real_name;
ALTER TABLE t_user RENAME COLUMN f_age TO __f_age;
ALTER TABLE t_user ADD COLUMN f_age INTEGER NOT NULL DEFAULT '0';
UPDATE t_user SET f_age = __f_age;
ALTER TABLE t_user DROP COLUMN __f_age;
CREATE UNIQUE INDEX t_user_i_age ON t_user (f_age);
CREATE UNIQUE INDEX t_user_i_name ON t_user (f_real_name);
`))
	})
}
