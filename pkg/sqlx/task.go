package sqlx

import (
	"context"
	"fmt"
	"runtime/debug"
)

type Task func(db DBExecutor) error

func (task Task) Run(db DBExecutor) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %s; calltrace:%s", fmt.Sprint(e), string(debug.Stack()))
		}
	}()
	return task(db)
}

func NewTasks(db DBExecutor) *Tasks {
	return &Tasks{
		db: db,
	}
}

type Tasks struct {
	db    DBExecutor
	tasks []Task
}

func (tasks Tasks) With(task ...Task) *Tasks {
	tasks.tasks = append(tasks.tasks, task...)
	return &tasks
}

func (tasks *Tasks) Do() (err error) {
	if len(tasks.tasks) == 0 {
		return nil
	}

	return tasks.db.Transaction(tasks.db.Context(), func(ctx context.Context) error {
		for i := range tasks.tasks {
			if err := tasks.tasks[i].Run(tasks.db.WithContext(ctx)); err != nil {
				return err
			}
		}
		return nil
	})
}
