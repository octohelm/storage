package dal

import "database/sql/driver"

type ModelWithAutoIncrement interface {
	SetAutoIncrementID(u int64)
}

type ModelWithSoftDelete interface {
	SoftDeleteFieldAndZeroValue() (string, driver.Value)
	SetSoftDelete() driver.Value
}
