package sqlite

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/octohelm/storage/internal/sql/scanner"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/x/testing/bdd"
	"golang.org/x/sync/errgroup"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/migrator"
	sqlbuildercatalog "github.com/octohelm/storage/pkg/sqlbuilder/catalog"
	"github.com/octohelm/storage/testdata/model"
)

func NewAdapter(t testing.TB) adapter.Adapter {
	t.Helper()

	dir := t.TempDir()

	ctx := testutil.NewContext(t)

	u, _ := url.Parse(fmt.Sprintf("sqlite://%s", filepath.Join(dir, "sqlite.db")))

	a, err := Open(ctx, u)
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		_ = a.Close()
		_ = os.RemoveAll(dir)
	})

	return a
}

func TestCatalog(t *testing.T) {
	a := NewAdapter(t)

	bdd.FromT(t).Given("a db", func(b bdd.T) {
		ctx := testutil.NewContext(t)

		tables, err := a.Catalog(ctx)
		b.Then("could got catalog",
			bdd.NoError(err),
		)

		for table := range tables.Tables() {
			fmt.Println(table.TableName())
		}
	})
}

func TestMigrate(t *testing.T) {
	adt := NewAdapter(t)

	bdd.FromT(t).Given("a db", func(b bdd.T) {
		ctx := testutil.NewContext(t)

		b.When("do migrate", func(b bdd.T) {
			v1 := sqlbuildercatalog.From(&model.User{})

			b.Then("success",
				bdd.NoError(migrator.Migrate(ctx, adt, v1)),
			)

			b.When("do migrate v2", func(b bdd.T) {
				v2 := sqlbuildercatalog.From(&model.UserV2{})

				b.Then("success",
					bdd.NoError(migrator.Migrate(ctx, adt, v2)),
				)

				b.When("rollback", func(b bdd.T) {
					b.Then("success",
						bdd.NoError(migrator.Migrate(ctx, adt, v1)),
					)
				})
			})
		})
	})
}

func TestParallel(t *testing.T) {
	adt := NewAdapter(t)

	bdd.FromT(t).Given("a db", func(b bdd.T) {
		ctx := testutil.NewContext(t)

		b.When("do migrate", func(b bdd.T) {
			v1 := sqlbuildercatalog.From(&model.User{})

			b.Then("success",
				bdd.NoError(migrator.Migrate(ctx, adt, v1)),
			)

			eg := &errgroup.Group{}

			for i := range 500 {
				eg.Go(func() error {
					_, err := adt.Exec(ctx, sqlfrag.Pair(
						"INSERT INTO t_user (f_name, f_age) VALUES (?,?)",
						fmt.Sprintf("name_%d", i), i+10,
					))
					return err
				})

				eg.Go(func() error {
					rows, err := adt.Query(ctx, sqlfrag.Pair("SELECT * FROM t_user"))
					if err != nil {
						return err
					}

					userList := make([]model.User, 0)
					_ = scanner.Scan(ctx, rows, &userList)

					return nil
				})
			}

			b.Then("insert & query without errors",
				bdd.NoError(eg.Wait()),
			)
		})
	})
}
