package time

import (
	"database/sql/driver"
	"time"

	"github.com/octohelm/storage/pkg/sqltype"
)

var _ sqltype.WithCreationTime = &CreationTime{}

type CreationTime struct {
	// 创建时间
	CreatedAt Timestamp `db:"f_created_at,default='0'" json:"createdAt"`
}

func (times *CreationTime) MarkCreatedAt() {
	if times.CreatedAt.IsZero() {
		times.CreatedAt = Timestamp(time.Now())
	}
}

var _ sqltype.WithModificationTime = &CreationUpdationTime{}

type (
	CreationUpdationTime     = CreationModificationTime
	CreationModificationTime struct {
		CreationTime
		// 更新时间
		UpdatedAt Timestamp `db:"f_updated_at,default='0'" json:"updatedAt"`
	}
)

func (times *CreationUpdationTime) MarkModifiedAt() {
	if times.UpdatedAt.IsZero() {
		times.UpdatedAt = Timestamp(time.Now())
	}
}

func (times *CreationUpdationTime) MarkCreatedAt() {
	times.MarkModifiedAt()

	if times.CreatedAt.IsZero() {
		times.CreatedAt = times.UpdatedAt
	}
}

var _ sqltype.WithSoftDelete = &CreationUpdationDeletionTime{}

type CreationUpdationDeletionTime struct {
	CreationUpdationTime
	// 删除时间
	DeletedAt Timestamp `db:"f_deleted_at,default='0'" json:"deletedAt,omitempty"`
}

func (CreationUpdationDeletionTime) SoftDeleteFieldAndZeroValue() (string, driver.Value) {
	return "DeletedAt", int64(0)
}

func (times *CreationUpdationDeletionTime) MarkDeletedAt() {
	times.MarkModifiedAt()

	times.DeletedAt = times.UpdatedAt
}
