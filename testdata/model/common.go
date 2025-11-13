package model

import (
	"database/sql/driver"
	"time"

	"github.com/octohelm/storage/pkg/sqltype"
)

type OperateTime struct {
	CreatedAt Datetime `db:"f_created_at,default=CURRENT_TIMESTAMP,onupdate=CURRENT_TIMESTAMP"`
	UpdatedAt int64    `db:"f_updated_at,default=0"`
}

func (v *OperateTimeWithDeleted) MarkCreatedAt() {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = Datetime(time.Now())
	}
}

type OperateTimeWithDeleted struct {
	OperateTime
	DeletedAt int64 `db:"f_deleted_at,default=0"`
}

var _ sqltype.WithSoftDelete = &OperateTimeWithDeleted{}

func (v OperateTimeWithDeleted) SoftDeleteFieldAndZeroValue() (string, driver.Value) {
	return "DeletedAt", int64(v.DeletedAt)
}

var _ sqltype.DeletedAtMarker = &OperateTimeWithDeleted{}

func (v *OperateTimeWithDeleted) MarkDeletedAt() {
	v.DeletedAt = time.Now().Unix()
}
