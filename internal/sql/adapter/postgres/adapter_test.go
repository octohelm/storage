package postgres

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/octohelm/storage/internal/sql/adapter"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/migrator"
	sqlbuildercatalog "github.com/octohelm/storage/pkg/sqlbuilder/catalog"
	"github.com/octohelm/storage/testdata/model"
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

func Test(t *testing.T) {
	a := NewAdapter(t)

	t.Run("#Catalog", func(t *testing.T) {
		ctx := testutil.NewContext(t)
		tables, err := a.Catalog(ctx)
		testutil.Expect(t, err, testutil.Be[error](nil))
		spew.Dump(slices.Sorted(sqlbuilder.TableNames(tables)))
	})
}

func TestMigrate(t *testing.T) {
	a := NewAdapter(t)

	t.Run("Create Catalog", func(t *testing.T) {
		ctx := testutil.NewContext(t)

		cat := sqlbuildercatalog.From(&model.User{})
		err := migrator.Migrate(ctx, a, cat)
		testutil.Expect(t, err, testutil.Be[error](nil))

		t.Run("Migrate To TableV2", func(t *testing.T) {
			catV2 := sqlbuildercatalog.From(&model.UserV2{})

			err := migrator.Migrate(ctx, a, catV2)
			testutil.Expect(t, err, testutil.Be[error](nil))

			t.Run("Rollback", func(t *testing.T) {
				err := migrator.Migrate(ctx, a, cat)
				testutil.Expect(t, err, testutil.Be[error](nil))
			})
		})
	})
}
