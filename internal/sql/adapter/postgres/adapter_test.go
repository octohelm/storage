package postgres

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/migrator"
	sqlbuildercatalog "github.com/octohelm/storage/pkg/sqlbuilder/catalog"
	"github.com/octohelm/storage/testdata/model"
	"github.com/octohelm/x/testing/bdd"
)

func NewAdapter(t testing.TB) adapter.Adapter {
	t.Helper()

	ctx := testutil.NewContext(t)

	u, _ := url.Parse("postgres://postgres@localhost/t_" + strconv.FormatInt(time.Now().Unix(), 10) + "?sslmode=disable&pool_max_conns=10")

	a, err := Open(ctx, u)
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		a.Close()
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
