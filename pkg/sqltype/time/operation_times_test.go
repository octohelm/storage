package time

import (
	"database/sql/driver"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestOperationTimes(t *testing.T) {
	created := &CreationTime{}
	created.MarkCreatedAt()
	createdAt := created.CreatedAt
	created.MarkCreatedAt()

	updated := &CreationUpdationTime{}
	updated.MarkCreatedAt()
	updatedCreatedAt := updated.CreatedAt
	updatedAt := updated.UpdatedAt
	updated.MarkModifiedAt()

	deleted := &CreationUpdationDeletionTime{}
	deleted.MarkDeletedAt()
	fieldName, zeroValue := deleted.SoftDeleteFieldAndZeroValue()

	Then(t, "操作时间结构体只在零值时初始化创建时间",
		Expect(createdAt.IsZero(), Equal(false)),
		Expect(created.CreatedAt, Equal(createdAt)),
	)

	Then(t, "创建更新时间会同步初始化 UpdatedAt 且不覆盖已有 CreatedAt",
		Expect(updatedCreatedAt.IsZero(), Equal(false)),
		Expect(updatedAt.IsZero(), Equal(false)),
		Expect(updated.CreatedAt, Equal(updatedCreatedAt)),
	)

	Then(t, "删除时间沿用最近一次更新时间并暴露软删字段定义",
		Expect(deleted.DeletedAt.IsZero(), Equal(false)),
		Expect(deleted.DeletedAt, Equal(deleted.UpdatedAt)),
		Expect(fieldName, Equal("DeletedAt")),
		Expect(zeroValue, Equal(driver.Value(int64(0)))),
	)
}
