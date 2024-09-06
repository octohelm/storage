package dal_test

import (
	"context"

	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/dal"
	"github.com/octohelm/storage/pkg/dal/compose"
	"github.com/octohelm/storage/pkg/datatypes"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/filter"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/testdata/model"
)

type FilterUserByAge struct {
	Age *filter.Filter[int64] `name:"age" in:"query"`
}

func (i FilterUserByAge) Apply(q dal.Querier) dal.Querier {
	return compose.ApplyQuerierFromFilter(q, model.UserT.Age, i.Age)
}

func TestCRUD(t *testing.T) {
	ctxs := []context.Context{
		ContextWithDatabase(t, "dal_sql_crud", ""),
		ContextWithDatabase(t, "dal_sql_crud", "postgres://postgres@localhost?sslmode=disable"),
	}

	for i := range ctxs {
		ctx := ctxs[i]

		t.Run("batch insert", func(t *testing.T) {
			err := dal.Prepare(&model.User{}).
				Values(
					func(yield func(*model.User) bool) {
						for i := 0; i < 100; i++ {
							usr := &model.User{
								Name: uuid.New().String(),
								Age:  int64(i),
							}
							if !yield(usr) {
								return
							}
						}
					},
					model.UserT.Name,
					model.UserT.Age,
				).
				OnConflict(model.UserT.I.IName).
				DoNothing().
				Save(ctx)
			testutil.Expect(t, err, testutil.Be[error](nil))

			c, err := dal.From(model.UserT).Count(ctx)
			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, c, testutil.Be(100))

			err = dal.Prepare(&model.User{}).ForDelete(dal.HardDelete()).
				Where(model.UserT.Age.V(sqlbuilder.Gte(int64(0)))).Save(ctx)
			testutil.Expect(t, err, testutil.Be[error](nil))
		})

		t.Run("Save one user", func(t *testing.T) {
			usr := &model.User{
				Name: uuid.New().String(),
				Age:  100,
			}
			err := dal.Prepare(usr).IncludesZero(model.UserT.Nickname).
				Returning(model.UserT.ID).Scan(usr).
				Save(ctx)

			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, usr.ID, testutil.Not(testutil.Be(uint64(0))))

			t.Run("Save same user agent, should conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name: usr.Name,
				}
				err := dal.Prepare(usr2).Save(ctx)
				testutil.Expect(t, dberr.IsErrConflict(err), testutil.Be(true))
			})

			t.Run("Save same user again, when set ignore should not clause conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name:     usr.Name,
					Nickname: "test",
				}

				err := dal.Prepare(usr2).
					OnConflict(model.UserT.I.IName).DoNothing().
					Returning(model.UserT.ID, model.UserT.Age).Scan(usr2).
					Save(ctx)

				testutil.Expect(t, err, testutil.Be[error](nil))
			})

			t.Run("Save same user again, when set ignore should not clause conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name:     usr.Name,
					Nickname: "test",
				}

				err := dal.Prepare(usr2).
					OnConflict(model.UserT.I.IName).DoUpdateSet(model.UserT.Nickname).
					Returning(model.UserT.ID, model.UserT.Age, model.UserT.Username).Scan(usr2).
					Save(ctx)

				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, usr2.ID, testutil.Be(usr.ID))
				testutil.Expect(t, usr2.Age, testutil.Be(usr.Age))
			})

			t.Run("Update", func(t *testing.T) {
				usr2 := &model.User{
					Nickname: "test test",
				}
				update := dal.Prepare(usr2).Where(model.UserT.ID.V(sqlbuilder.Eq[uint64](100)))

				err := update.Save(ctx)
				testutil.Expect(t, err, testutil.Be[error](nil))
			})

			t.Run("SoftDelete", func(t *testing.T) {
				deletedUser := &model.User{}
				update := dal.Prepare(&model.User{}).ForDelete().
					Returning().Scan(deletedUser).
					Where(model.UserT.ID.V(sqlbuilder.Eq(usr.ID)))

				err := update.Save(ctx)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
			})

			t.Run("Delete", func(t *testing.T) {
				deletedUser := &model.User{}

				update := dal.Prepare(&model.User{}).ForDelete(dal.HardDelete()).
					Returning().Scan(deletedUser).
					Where(model.UserT.ID.V(sqlbuilder.Eq(usr.ID)))

				err := update.Save(ctx)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
			})
		})

		t.Run("Insert multi Users and Orgs", func(t *testing.T) {
			err := dal.Tx(ctx, &model.Org{}, func(ctx context.Context) error {
				for i := 0; i < 2; i++ {
					org := &model.Org{
						Name: uuid.New().String(),
					}
					if err := dal.Prepare(org).Returning(model.OrgT.ID).Scan(org).Save(ctx); err != nil {
						return err
					}
				}

				for i := 0; i < 110; i++ {
					usr := &model.User{
						Name: uuid.New().String(),
						Age:  int64(i),
					}

					err := dal.Prepare(usr).IncludesZero(model.UserT.Nickname).
						Returning(model.UserT.ID).Scan(usr).
						Save(ctx)
					if err != nil {
						return err
					}

					if i >= 100 {
						if err := dal.Prepare(usr).ForDelete().Where(
							model.UserT.Age.V(sqlbuilder.Eq[int64](usr.Age)),
						).Save(ctx); err != nil {
							return err
						}
					}

					orgUsr := &model.OrgUser{
						UserID: usr.ID,
						OrgID:  usr.ID%2 + 1,
					}
					if err := dal.Prepare(orgUsr).Save(ctx); err != nil {
						return err
					}
				}

				return nil
			})

			testutil.Expect(t, err, testutil.Be[error](nil))

			t.Run("Then Queries", func(t *testing.T) {
				t.Run("Count", func(t *testing.T) {
					c, err := dal.From(model.UserT).Count(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, c, testutil.Be(100))
				})

				t.Run("List all", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(100))
				})

				t.Run("List partial with cancel", func(t *testing.T) {
					users := make([]*model.User, 0)

					ctx, cancel := context.WithCancel(ctx)

					err := dal.From(model.UserT).
						Scan(dal.Recv(func(user *model.User) error {
							users = append(users, user)

							if len(users) >= 10 {
								cancel()
							}

							return nil
						})).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
				})

				t.Run("List all", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT, dal.IncludeAllRecord()).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(110))
				})

				t.Run("List all limit 10", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT).
						Limit(10).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
				})

				t.Run("List all offset limit 10", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT).
						Offset(10).Limit(10).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
					testutil.Expect(t, users[0].ID > 1, testutil.Be(true))
				})

				t.Run("List desc order by", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT).
						OrderBy(sqlbuilder.DescOrder(model.UserT.ID)).
						Offset(10).Limit(10).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
					testutil.Expect(t, users[0].ID > users[1].ID, testutil.Be(true))
				})

				t.Run("List where", func(t *testing.T) {
					users := make([]model.User, 0)

					err := dal.From(model.UserT).
						Apply(FilterUserByAge{
							Age: filter.Eq(int64(10)),
						}).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(1))
				})

				t.Run("List where with in", func(t *testing.T) {
					orgUsers := make([]model.OrgUser, 0)

					err := dal.From(model.OrgUserT).
						Where(model.OrgUserT.UserID.V(dal.InSelect(
							model.UserT.ID,
							dal.From(model.UserT).Where(model.UserT.Age.V(sqlbuilder.Eq(int64(10)))),
						))).
						Scan(&orgUsers).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(orgUsers), testutil.Be(1))
				})

				t.Run("List where join", func(t *testing.T) {
					users := make([]struct {
						model.User
						Org model.Org
					}, 0)

					err := dal.From(model.UserT).
						Join(model.OrgUserT, model.OrgUserT.UserID.V(sqlbuilder.EqCol(model.UserT.ID))).
						Join(model.OrgT, model.OrgT.ID.V(sqlbuilder.EqCol(model.OrgUserT.OrgID))).
						Where(model.UserT.Age.V(sqlbuilder.Eq(int64(10)))).
						Scan(&users).
						Find(ctx)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(1))
					testutil.Expect(t, users[0].Org.Name, testutil.Not(testutil.Be("")))
				})
			})
		})
	}
}

func TestMultipleTxLockedWithSqlite(t *testing.T) {
	ctx := ContextWithDatabase(t, "sql_test", "")

	t.Run("concurrent insert && query", func(t *testing.T) {
		usr2 := &model.User{
			Name:     "test",
			Nickname: "test",
		}

		wg := &sync.WaitGroup{}

		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := dal.Prepare(usr2).
					OnConflict(model.UserT.I.IName).DoUpdateSet(model.UserT.Nickname).
					Save(ctx)

				testutil.Expect(t, err, testutil.Be[error](nil))
			}()
		}

		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				//err := Tx(ctx, usr2, func(ctx context.Context) error {
				//	return dal.Prepare(usr2).
				//		OnConflict(model.UserT.I.IName).DoUpdateSet(model.UserT.Nickname).
				//		Save(ctx)
				//})
				//
				//testutil.Expect(t, err, testutil.Be[error](nil))
			}()
		}

		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := dal.From(model.UserT).
					Scan(dal.Recv(func(v *model.User) error {
						return nil
					})).
					Find(ctx)
				testutil.Expect(t, err, testutil.Be[error](nil))
			}()
		}

		wg.Wait()
	})
}

func ContextWithDatabase(t testing.TB, name string, endpoint string) context.Context {
	t.Helper()
	ctx := testutil.NewContext(t)

	cat := &sqlbuilder.Tables{}
	cat.Add(model.UserT)
	cat.Add(model.OrgT)
	cat.Add(model.OrgUserT)

	db := &dal.Database{
		EnableMigrate: true,
	}

	if endpoint != "" {
		db.Endpoint = *must(datatypes.ParseEndpoint(endpoint))
	}

	db.ApplyCatalog(name, cat)
	db.SetDefaults()
	err := db.Init(ctx)
	testutil.Expect(t, err, testutil.Be[error](nil))

	ctx = db.InjectContext(ctx)

	err = db.Run(ctx)
	testutil.Expect(t, err, testutil.Be[error](nil))

	t.Cleanup(func() {
		a := dal.SessionFor(ctx, name).Adapter()

		cat.Range(func(table sqlbuilder.Table, idx int) bool {
			_, e := a.Exec(ctx, a.Dialect().DropTable(table))
			testutil.Expect(t, e, testutil.Be[error](nil))
			return true
		})

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
