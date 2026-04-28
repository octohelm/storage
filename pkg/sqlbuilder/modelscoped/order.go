package modelscoped

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

// AscOrder 为类型化列创建升序排序表达式。
func AscOrder[Model internal.Model](col Column[Model]) Order[Model] {
	return sqlbuilder.AscOrder(col)
}

// DescOrder 为类型化列创建降序排序表达式。
func DescOrder[Model internal.Model](col Column[Model]) Order[Model] {
	return sqlbuilder.DescOrder(col)
}

// Order 是 sqlbuilder.Order 的类型化包装。
type Order[Model internal.Model] interface {
	sqlbuilder.Order
}
