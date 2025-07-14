package sqltype

import (
	"database/sql/driver"
	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type WithSoftDelete interface {
	SoftDeleteFieldAndZeroValue() (string, driver.Value)
}

type DeletedAtMarker interface {
	MarkDeletedAt()
}

type SoftDeleteValueGetter interface {
	GetDeletedAt() driver.Value
}

func HasSoftDelete[M sqlbuilder.Model]() bool {
	_, ok := any(new(M)).(WithSoftDelete)
	return ok
}
