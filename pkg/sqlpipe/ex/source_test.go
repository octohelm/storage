package ex

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/session"
	sessiondb "github.com/octohelm/storage/pkg/session/db"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlpipe"
	"github.com/octohelm/storage/testdata/model"
	modelfilter "github.com/octohelm/storage/testdata/model/filter/v2"
)

type Repo struct {
	User Executor[model.User]
}

func TestSourceExecutor(t *testing.T) {
	repo := &Repo{}

	for _, ctx := range []context.Context{
		ContextWithDatabase(t, "sqlpipe_crud", ""),
		ContextWithDatabase(t, "sqlpipe_crud", "postgres://postgres@localhost?sslmode=disable"),
	} {
		t.Run("batch insert", func(t *testing.T) {
			values := sqlpipe.Values(slices.Collect(func(yield func(*model.User) bool) {
				for i := 0; i < 100; i++ {
					usr := &model.User{
						Name: uuid.New().String(),
						Age:  int64(i),
					}
					if !yield(usr) {
						return
					}
				}
			}))

			users := make([]*model.User, 0)

			ex := FromSource(values)
			testutil.Expect(t, ex.Range(ctx, func(u *model.User) {
				users = append(users, u)
			}), testutil.Be[error](nil))

			testutil.Expect(t, len(users), testutil.Be(100))

			t.Run("re insert", func(t *testing.T) {
				ex := FromSource(sqlpipe.Values(users)).PipeE(
					sqlpipe.OnConflictDoNothing(model.UserT.I.IName),
				)

				updatedUsers := make([]*model.User, 0)
				testutil.Expect(t, ex.Range(ctx, func(u *model.User) {
					updatedUsers = append(updatedUsers, u)
				}), testutil.Be[error](nil))

				testutil.Expect(t, len(updatedUsers), testutil.Be(100))
			})

			t.Run("then could count", func(t *testing.T) {
				var count int64
				err := repo.User.CountTo(ctx, &count)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, count, testutil.Be[int64](100))
			})

			t.Run("then could list all", func(t *testing.T) {
				items, err := repo.User.List(ctx)

				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, len(items), testutil.Be(100))
			})

			t.Run("then could list filter", func(t *testing.T) {
				items, err := repo.User.PipeE(
					&modelfilter.UserByAge{
						Age: filter.Lt[int64](50),
					},
					sqlpipe.Limit[model.User](10),
				).List(ctx)

				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, len(items), testutil.Be(10))
			})

			t.Run("then could filtered", func(t *testing.T) {
				var count int64

				err := repo.User.PipeE(
					&modelfilter.UserByAge{
						Age: filter.Lt[int64](50),
					},
					sqlpipe.Limit[model.User](10),
				).CountTo(ctx, &count)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, count, testutil.Be[int64](50))
			})

			t.Run("then delete", func(t *testing.T) {
				t.Run("delete one ", func(t *testing.T) {
					src := repo.User.PipeE(
						&modelfilter.UserByAge{
							Age: filter.Eq[int64](0),
						},
						sqlpipe.DoDelete[model.User](),
					)

					deleted, err := src.FindOne(ctx)
					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, deleted.Age, testutil.Be[int64](0))

					t.Run("then could count", func(t *testing.T) {
						var count int64

						err := repo.User.CountTo(ctx, &count)
						testutil.Expect(t, err, testutil.Be[error](nil))
						testutil.Expect(t, count, testutil.Be[int64](99))
					})
				})

				t.Run("should got empty data when all deleted", func(t *testing.T) {
					src := repo.User.PipeE(
						sqlpipe.DoDelete[model.User](),
					)

					testutil.Expect(t, src.Commit(ctx), testutil.Be[error](nil))

					t.Run("then could count", func(t *testing.T) {
						var count int64

						err := repo.User.CountTo(ctx, &count)
						testutil.Expect(t, err, testutil.Be[error](nil))
						testutil.Expect(t, count, testutil.Be[int64](0))
					})
				})
			})
		})
	}
}

func ContextWithDatabase(t testing.TB, name string, endpoint string) context.Context {
	t.Helper()
	ctx := testutil.NewContext(t)

	cat := &sqlbuilder.Tables{}
	cat.Add(model.UserT)
	cat.Add(model.OrgT)
	cat.Add(model.OrgUserT)

	db := &sessiondb.Database{
		EnableMigrate: true,
	}

	if endpoint != "" {
		db.Endpoint = *must(sessiondb.ParseEndpoint(endpoint))
	}

	db.ApplyCatalog(name, cat)
	db.SetDefaults()
	err := db.Init(ctx)
	testutil.Expect(t, err, testutil.Be[error](nil))

	ctx = db.InjectContext(ctx)

	err = db.Run(ctx)
	testutil.Expect(t, err, testutil.Be[error](nil))

	t.Cleanup(func() {
		a := session.For(ctx, name).Adapter()

		for table := range cat.Tables() {
			_, e := a.Exec(ctx, a.Dialect().DropTable(table))
			testutil.Expect(t, e, testutil.Be[error](nil))
		}

		err := a.Close()
		testutil.Expect(t, err, testutil.Be[error](nil))
	})

	return ctx
}

func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
