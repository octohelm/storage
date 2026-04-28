package sqltype

import (
	"database/sql/driver"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

// WithSoftDelete 表示模型暴露软删除字段及其零值。
type WithSoftDelete interface {
	SoftDeleteFieldAndZeroValue() (string, driver.Value)
}

// DeletedAtMarker 表示模型支持标记删除时间。
type DeletedAtMarker interface {
	MarkDeletedAt()
}

// SoftDeleteValueGetter 表示模型可读取当前软删除值。
type SoftDeleteValueGetter interface {
	GetDeletedAt() driver.Value
}

// HasSoftDelete 判断模型是否实现软删除能力。
func HasSoftDelete[M sqlbuilder.Model]() bool {
	_, ok := any(new(M)).(WithSoftDelete)
	return ok
}
