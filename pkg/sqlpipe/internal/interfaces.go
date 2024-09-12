package internal

import "database/sql/driver"

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
