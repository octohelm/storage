package sqlx_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/dberr"
	"github.com/octohelm/storage/pkg/migrator"
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlx"
	"github.com/octohelm/storage/testdata/model"
)

func TestCRUD(t *testing.T) {
	dbs := []sqlx.DBExecutor{
		NewPostgresDatabase(t, "test_crud"),
		NewSqliteDatabase(t, "test_crud"),
	}

	for i := range dbs {
		db := dbs[i]

		t.Run(db.DriverName(), func(t *testing.T) {
			t.Run("Insert single", func(t *testing.T) {
				user := model.User{
					Name:   uuid.New().String(),
					Gender: model.GENDER__MALE,
				}

				_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
				testutil.Expect(t, err, testutil.Be[error](nil))

				t.Run("update", func(t *testing.T) {
					user.Gender = model.GENDER__FEMALE
					_, err := db.ExecExpr(
						sqlbuilder.Update(db.T(&user)).
							Set(sqlx.AsAssignments(db, &user)...).
							Where(
								db.T(&user).F("Name").Eq(user.Name),
							),
					)
					testutil.Expect(t, err, testutil.Be[error](nil))
				})

				t.Run("select", func(t *testing.T) {
					userForSelect := model.User{}

					err := db.QueryExprAndScan(
						sqlbuilder.Select(nil).From(
							db.T(&user),
							sqlbuilder.Where(db.T(&user).F("Name").Eq(user.Name)),
							sqlbuilder.Comment("FindUser"),
						),
						&userForSelect,
					)

					testutil.Expect(t, err, testutil.Be[error](nil))
					testutil.Expect(t, user.Name, testutil.Equal(userForSelect.Name))
					testutil.Expect(t, user.Gender, testutil.Equal(userForSelect.Gender))

					t.Run("conflict", func(t *testing.T) {
						_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
						testutil.Expect(t, sqlx.DBErr(err).IsConflict(), testutil.Be(true))
					})

					t.Run("Delete", func(t *testing.T) {
						_, err := db.ExecExpr(
							sqlbuilder.Delete().From(
								db.T(&user),
								sqlbuilder.Where(db.T(&user).F("Name").Eq(user.Name)),
							),
						)
						testutil.Expect(t, err, testutil.Be[error](nil))

						t.Run("Re select", func(t *testing.T) {
							userForSelectV2 := model.User{}

							err := db.QueryExprAndScan(
								sqlbuilder.Select(nil).From(
									db.T(&user),
									sqlbuilder.Where(db.T(&user).F("Name").Eq(user.Name)),
								),
								&userForSelectV2,
							)

							testutil.Expect(t, dberr.IsErrNotFound(err), testutil.Be(true))
						})
					})
				})

				t.Run("cancel", func(t *testing.T) {
					c := testutil.NewContext(t)
					ctx, cancel := context.WithCancel(c)
					db2 := db.WithContext(ctx)

					go func() {
						time.Sleep(3 * time.Millisecond)
						cancel()
					}()

					err := sqlx.NewTasks(db2).
						With(
							func(db sqlx.DBExecutor) error {
								stmt := sqlx.InsertToDB(db,
									&user, nil,
									sqlbuilder.OnConflict(model.UserT.I.IName).DoNothing(),
								)
								_, err := db.ExecExpr(stmt)
								return err
							},
							func(db sqlx.DBExecutor) error {
								time.Sleep(200 * time.Millisecond)
								return nil
							}).
						Do()
					testutil.Expect(t, err, testutil.Not(testutil.Be[error](nil)))
				})
			})
		})
	}
}

func TestQueries(t *testing.T) {
	dbs := []sqlx.DBExecutor{
		NewPostgresDatabase(t, "test_crud"),
		NewSqliteDatabase(t, "test_crud"),
	}

	for i := range dbs {
		db := dbs[i]

		{
			columns := db.T(&model.User{}).Cols("Name", "Gender", "Age")
			values := make([]interface{}, 0)

			for i := 0; i < 1000; i++ {
				values = append(values, uuid.New().String(), model.GENDER__MALE, i+1)
			}

			_, err := db.ExecExpr(sqlbuilder.Insert().Into(db.T(&model.User{})).Values(columns, values...))
			testutil.Expect(t, err, testutil.Be[error](nil))
		}

		t.Run("select to slice", func(t *testing.T) {
			users := make([]model.User, 0)
			err := db.QueryExprAndScan(
				sqlbuilder.Select(nil).From(
					db.T(&model.User{}),
					sqlbuilder.Where(db.T(&model.User{}).F("Gender").Eq(model.GENDER__MALE))),
				&users,
			)
			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, users, testutil.HaveLen[[]model.User](1000))
		})

		t.Run("select to set", func(t *testing.T) {
			userSet := UserSet{}
			err := db.QueryExprAndScan(
				sqlbuilder.Select(nil).From(db.T(&model.User{}), sqlbuilder.Where(db.T(&model.User{}).F("Gender").Eq(model.GENDER__MALE))),
				userSet,
			)
			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, userSet, testutil.HaveLen[UserSet](1000))
		})

		t.Run("not found", func(t *testing.T) {
			user := model.User{}
			err := db.QueryExprAndScan(
				sqlbuilder.Select(nil).From(
					db.T(&model.User{}),
					sqlbuilder.Where(db.T(&model.User{}).F("ID").Eq(1001)),
				),
				&user,
			)
			testutil.Expect(t, sqlx.DBErr(err).IsNotFound(), testutil.Be(true))
		})

		t.Run("count", func(t *testing.T) {
			count := 0
			err := db.QueryExprAndScan(
				sqlbuilder.Select(sqlbuilder.Count()).From(db.T(&model.User{})),
				&count,
			)
			testutil.Expect(t, err, testutil.Be[error](nil))
			testutil.Expect(t, count, testutil.Equal(1000))
		})

		t.Run("canceled", func(t *testing.T) {
			ctx, cancel := context.WithCancel(db.Context())
			db2 := db.WithContext(ctx)

			go func() {
				time.Sleep(2 * time.Millisecond)
				cancel()
			}()

			userSet := UserSet{}
			err := db2.QueryExprAndScan(
				sqlbuilder.Select(nil).From(db.T(&model.User{}),
					sqlbuilder.Where(db.T(&model.User{}).F("Gender").Eq(model.GENDER__MALE)),
				),
				userSet,
			)

			testutil.Expect(t, err, testutil.Not(testutil.Be[error](nil)))
		})
	}
}

type UserSet map[string]*model.User

func (UserSet) New() interface{} {
	return &model.User{}
}

func (u UserSet) Next(v interface{}) error {
	user := v.(*model.User)
	u[user.Name] = user
	return nil
}

func NewPostgresDatabase(t testing.TB, name string) sqlx.DBExecutor {
	t.Helper()
	ctx := testutil.NewContext(t)

	dbTest := sqlx.NewDatabase(name)
	dbTest.Register(&model.User{})

	db, err := dbTest.OpenDB(ctx, "postgres://postgres@localhost?sslmode=disable")
	testutil.Expect(t, err, testutil.Be[error](nil))

	err = migrator.Migrate(ctx, db, &dbTest.Tables)
	testutil.Expect(t, err, testutil.Be[error](nil))

	db = db.WithContext(ctx)

	t.Cleanup(func() {
		dbTest.Tables.Range(func(table sqlbuilder.Table, idx int) bool {
			_, e := db.Exec(ctx, db.Dialect().DropTable(table))
			testutil.Expect(t, e, testutil.Be[error](nil))
			return true
		})
		err = db.Close()
		testutil.Expect(t, err, testutil.Be[error](nil))
	})

	return db
}

func NewSqliteDatabase(t testing.TB, name string) sqlx.DBExecutor {
	t.Helper()
	ctx := testutil.NewContext(t)

	dbTest := sqlx.NewDatabase(name)
	dbTest.Register(&model.User{})

	dir := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	s := fmt.Sprintf("sqlite://%s", filepath.Join(dir, "sqlite.db"))
	fmt.Println(s)

	db, err := dbTest.OpenDB(ctx, s)
	testutil.Expect(t, err, testutil.Be[error](nil))

	err = migrator.Migrate(ctx, db, &dbTest.Tables)
	testutil.Expect(t, err, testutil.Be[error](nil))

	db = db.WithContext(ctx)

	t.Cleanup(func() {
		dbTest.Tables.Range(func(table sqlbuilder.Table, idx int) bool {
			_, e := db.Exec(ctx, db.Dialect().DropTable(table))
			testutil.Expect(t, e, testutil.Be[error](nil))
			return true
		})
		err = db.Close()
		testutil.Expect(t, err, testutil.Be[error](nil))
	})

	return db
}
