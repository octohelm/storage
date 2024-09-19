package sqltype

import "database/sql/driver"

type WithCreationTime interface {
	MarkCreatedAt()
}

type WithUpdationTime interface {
	MarkUpdatedAt()
}

type WithSoftDelete interface {
	SoftDeleteFieldAndZeroValue() (string, driver.Value)
}

type DeletedAtMarker interface {
	MarkDeletedAt()
}

type SoftDeleteValueGetter interface {
	GetDeletedAt() driver.Value
}
