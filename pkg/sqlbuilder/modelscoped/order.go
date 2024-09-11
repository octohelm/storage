package modelscoped

import (
	"github.com/octohelm/storage/pkg/sqlbuilder"
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

func AscOrder[Model internal.Model](col Column[Model]) Order[Model] {
	return sqlbuilder.AscOrder(col)
}

func DescOrder[Model internal.Model](col Column[Model]) Order[Model] {
	return sqlbuilder.DescOrder(col)
}

type Order[Model internal.Model] interface {
	sqlbuilder.Order
}
