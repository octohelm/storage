package model

import (
	"database/sql/driver"
	"time"

	"github.com/octohelm/storage/pkg/datatypes"
)

type OperateTime struct {
	CreatedAt datatypes.Datetime `db:"f_created_at,default=CURRENT_TIMESTAMP,onupdate=CURRENT_TIMESTAMP"`
	UpdatedAt int64              `db:"f_updated_at,default='0'"`
}

type OperateTimeWithDeleted struct {
	OperateTime
	DeletedAt int64 `db:"f_deleted_at,default='0'"`
}

func (v *OperateTimeWithDeleted) SoftDeleteFieldAndZeroValue() (string, driver.Value) {
	return "DeletedAt", int64(0)
}

func (v *OperateTimeWithDeleted) SetSoftDelete() driver.Value {
	v.DeletedAt = time.Now().Unix()
	return v.DeletedAt
}
