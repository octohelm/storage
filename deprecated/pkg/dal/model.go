package dal

import (
	"database/sql/driver"

	"github.com/octohelm/storage/pkg/sqlbuilder"
)

type ModelNewer interface {
	New() sqlbuilder.Model
}

type ModelWithAutoIncrement interface {
	SetAutoIncrementID(u int64)
}

type ModelWithCreationTime interface {
	MarkCreatedAt()
}

type ModelWithUpdationTime interface {
	MarkUpdatedAt()
}

type ModelWithSoftDelete interface {
	SoftDeleteFieldAndZeroValue() (string, driver.Value)
}

type DeletedAtMarker interface {
	MarkDeletedAt()
}

type SoftDeleteValueGetter interface {
	GetDeletedAt() driver.Value
}
