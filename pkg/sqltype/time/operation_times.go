package time

import (
	"database/sql/driver"
	"time"

	"github.com/octohelm/storage/pkg/sqltype"
)

var _ sqltype.WithCreationTime = &CreationTime{}

// CreationTime 定义带创建时间字段的通用结构。
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
	// CreationUpdationTime 是兼容旧命名的更新时间结构别名。
	CreationUpdationTime = CreationModificationTime
	// CreationModificationTime 定义同时包含创建和更新时间的通用结构。
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

// CreationUpdationDeletionTime 定义带软删除时间的通用结构。
type CreationUpdationDeletionTime struct {
	CreationUpdationTime
	// 删除时间
	DeletedAt Timestamp `db:"f_deleted_at,default='0'" json:"deletedAt"`
}

func (CreationUpdationDeletionTime) SoftDeleteFieldAndZeroValue() (string, driver.Value) {
	return "DeletedAt", int64(0)
}

func (times *CreationUpdationDeletionTime) MarkDeletedAt() {
	times.MarkModifiedAt()

	times.DeletedAt = times.UpdatedAt
}
