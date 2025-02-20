package model

import (
	"time"

	"database/sql/driver"
	"github.com/octohelm/storage/deprecated/pkg/datatypes"
)

type OperateTime struct {
	CreatedAt datatypes.Datetime `db:"f_created_at,default=CURRENT_TIMESTAMP,onupdate=CURRENT_TIMESTAMP"`
	UpdatedAt int64              `db:"f_updated_at,default='0'"`
}

func (v *OperateTimeWithDeleted) MarkCreatedAt() {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = datatypes.Datetime(time.Now())
	}
}

type OperateTimeWithDeleted struct {
	OperateTime
	DeletedAt int64 `db:"f_deleted_at,default='0'"`
}

func (v OperateTimeWithDeleted) SoftDeleteFieldAndZeroValue() (string, driver.Value) {
	return "DeletedAt", int64(v.DeletedAt)
}

func (v *OperateTimeWithDeleted) MarkDeletedAt() {
	v.DeletedAt = time.Now().Unix()
}
