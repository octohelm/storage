package sqlx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/octohelm/storage/internal/testutil"
	"github.com/octohelm/storage/pkg/sqlx"
	"github.com/octohelm/storage/testdata/model"
)

func TestWithTasks(t *testing.T) {
	dbs := []sqlx.DBExecutor{
		NewPostgresDatabase(t, "test_for_user_with_tasks"),
		NewSqliteDatabase(t, "test_for_user_with_tasks"),
	}

	for i := range dbs {
		db := dbs[i]

		t.Run("rollback on task err", func(t *testing.T) {
			taskList := sqlx.NewTasks(db)

			taskList = taskList.With(func(db sqlx.DBExecutor) error {
				user := model.User{
					Name:   "test",
					Gender: model.GENDER__MALE,
				}
				_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
				return err
			})

			taskList = taskList.With(func(db sqlx.DBExecutor) error {
				user := model.User{
					Name:   "test",
					Gender: model.GENDER__MALE,
				}
				_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
				return err
			})

			err := taskList.Do()
			testutil.Expect(t, err, testutil.Not(testutil.Be[error](nil)))
		})

		t.Run("transaction chain", func(t *testing.T) {
			taskList := sqlx.NewTasks(db)

			taskList = taskList.With(func(db sqlx.DBExecutor) error {
				user := model.User{
					Name:   uuid.New().String(),
					Gender: model.GENDER__MALE,
				}

				_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
				return err
			})

			taskList = taskList.With(func(db sqlx.DBExecutor) error {
				subTaskList := sqlx.NewTasks(db)

				subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
					user := model.User{
						Name:   uuid.New().String(),
						Gender: model.GENDER__MALE,
						Age:    1,
					}
					_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
					return err
				})

				subTaskList = subTaskList.With(func(db sqlx.DBExecutor) error {
					user := model.User{
						Name:   uuid.New().String(),
						Gender: model.GENDER__MALE,
						Age:    2,
					}
					_, err := db.ExecExpr(sqlx.InsertToDB(db, &user, nil))
					return err
				})

				return subTaskList.Do()
			})

			err := taskList.Do()
			testutil.Expect(t, err, testutil.Be[error](nil))
		})
	}
}
