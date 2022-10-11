package dal

import (
	"database/sql/driver"
)

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
	MarkDeletedAt()
	SoftDeleteFieldAndValue() (string, driver.Value)
}
