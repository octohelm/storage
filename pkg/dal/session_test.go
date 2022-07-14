package dal

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/testdata/model"
)

func TestCRUD(t *testing.T) {
	ctxs := []context.Context{
		ContextWithDatabase(t, "dal_sql_crud", ""),
		ContextWithDatabase(t, "dal_sql_crud", "postgres://postgres@localhost?sslmode=disable"),
	}

	for i := range ctxs {
		ctx := ctxs[i]

		db := FromContext(ctx, "dal_sql_crud")

		t.Run("Save one user", func(t *testing.T) {
			usr := &model.User{
				Name: uuid.New().String(),
				Age:  100,
			}
			err := Prepare(usr, model.UserT.Nickname).
				Returning(model.UserT.ID).Scan(usr).
				Save(ctx, db)

			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, usr.ID, testutil.Not(testutil.Be(uint64(0))))

			t.Run("Save same user agent, should conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name: usr.Name,
				}
				err := Prepare(usr2).Save(ctx, db)
				testutil.Expect(t, dberr.IsErrConflict(err), testutil.Be(true))
			})

			t.Run("Save same user again, when set ignore should not clause conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name:     usr.Name,
					Nickname: "test",
				}

				err := Prepare(usr2).
					OnConflict(model.UserT.I.IName).DoNothing().
					Returning(model.UserT.ID, model.UserT.Age).Scan(usr2).
					Save(ctx, db)

				testutil.Expect(t, err, testutil.Be[error](nil))
			})

			t.Run("Save same user again, when set ignore should not clause conflict", func(t *testing.T) {
				usr2 := &model.User{
					Name:     usr.Name,
					Nickname: "test",
				}

				err := Prepare(usr2).
					OnConflict(model.UserT.I.IName).DoUpdateSet(model.UserT.Nickname).
					Returning(model.UserT.ID, model.UserT.Age, model.UserT.Username).Scan(usr2).
					Save(ctx, db)

				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, usr2.ID, testutil.Be(usr.ID))
				testutil.Expect(t, usr2.Age, testutil.Be(usr.Age))
			})

			t.Run("Update", func(t *testing.T) {
				usr2 := &model.User{
					Nickname: "test test",
				}
				update := Prepare(usr2).
					Where(model.UserT.ID.Eq(100))

				err := update.Save(ctx, db)
				testutil.Expect(t, err, testutil.Be[error](nil))
			})

			t.Run("SoftDelete", func(t *testing.T) {
				deletedUser := &model.User{}
				update := Prepare(&model.User{}).ForDelete().
					Returning().Scan(deletedUser).
					Where(model.UserT.ID.Eq(usr.ID))

				err := update.Save(ctx, db)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
			})

			t.Run("Delete", func(t *testing.T) {
				deletedUser := &model.User{}

				update := Prepare(&model.User{}).ForDelete(HardDelete()).
					Returning().Scan(deletedUser).
					Where(model.UserT.ID.Eq(usr.ID))

				err := update.Save(ctx, db)
				testutil.Expect(t, err, testutil.Be[error](nil))
				testutil.Expect(t, deletedUser.ID, testutil.Be(usr.ID))
			})
		})

		t.Run("Insert multi Users and Orgs", func(t *testing.T) {
			err := db.Tx(ctx, func(ctx context.Context) error {
				for i := 0; i < 2; i++ {
					org := &model.Org{
						Name: uuid.New().String(),
					}
					err := Prepare(org).
						Returning(model.OrgT.ID).Scan(org).
						Save(ctx, db)
					if err != nil {
						return err
					}
				}

				for i := 0; i < 110; i++ {
					usr := &model.User{
						Name: uuid.New().String(),
						Age:  int64(i),
					}

					err := Prepare(usr, model.UserT.Nickname).
						Returning(model.UserT.ID).Scan(usr).
						Save(ctx, db)
					if err != nil {
						return err
					}

					if i >= 100 {
						if err := Prepare(usr).ForDelete().Where(model.UserT.Age.Eq(usr.Age)).Save(ctx, db); err != nil {
							return err
						}
					}

					orgUsr := &model.OrgUser{
						UserID: usr.ID,
						OrgID:  usr.ID%2 + 1,
					}
					if err := Prepare(orgUsr).Save(ctx, db); err != nil {
						return err
					}
				}

				return nil
			})
			testutil.Expect(t, err, testutil.Be[error](nil))

			t.Run("Then Queries", func(t *testing.T) {
				t.Run("Count", func(t *testing.T) {
					c, err := From(model.UserT).Count(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, c, testutil.Be(100))
				})

				t.Run("List all", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(100))
				})

				t.Run("List all", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT, IncludeAllRecord()).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(110))
				})

				t.Run("List all limit 10", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT).
						Limit(10).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
				})

				t.Run("List all offset limit 10", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT).
						Offset(10).Limit(10).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
					testutil.Expect(t, users[0].ID > 1, testutil.Be(true))
				})

				t.Run("List desc order by", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT).
						OrderBy(sqlbuilder.DescOrder(model.UserT.ID)).
						Offset(10).Limit(10).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(10))
					testutil.Expect(t, users[0].ID > users[1].ID, testutil.Be(true))
				})

				t.Run("List where", func(t *testing.T) {
					users := make([]model.User, 0)

					err := From(model.UserT).
						Where(model.UserT.Age.Eq(10)).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(1))
				})

				t.Run("List where with in", func(t *testing.T) {
					orgUsers := make([]model.OrgUser, 0)

					err := From(model.OrgUserT).
						Where(model.OrgUserT.UserID.In(
							From(model.UserT).Select(model.UserT.ID).Where(model.UserT.Age.Eq(10)),
						)).
						Scan(&orgUsers).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(orgUsers), testutil.Be(1))
				})

				t.Run("List where join", func(t *testing.T) {
					users := make([]struct {
						model.User
						Org model.Org
					}, 0)

					err := From(model.UserT).
						Join(model.OrgUserT, model.OrgUserT.UserID.Eq(model.UserT.ID)).
						Join(model.OrgT, model.OrgT.ID.Eq(model.OrgUserT.OrgID)).
						Where(model.UserT.Age.Eq(10)).
						Scan(&users).
						Find(ctx, db)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, len(users), testutil.Be(1))
					testutil.Expect(t, users[0].Org.Name, testutil.Not(testutil.Be("")))
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

	db := &Database{
		Name:     name,
		Endpoint: endpoint,
	}
	db.SetDefaults()
	db.Init()

	ctx = db.InjectContext(ctx)

	err := db.Migrate(ctx, cat)
	testutil.Expect(t, err, testutil.Be[error](nil))

	t.Cleanup(func() {
		a := FromContext(ctx, name).Adapter()

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
