package sqlbuilder

import (
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

type Model = internal.Model

type ModelNewer[M Model] internal.ModelNewer[M]

type DataTypeDescriber interface {
	DataType(driverName string) string
}

type WithTableDescription interface {
	TableDescription() []string
}

type Indexes map[string][]string

type WithPrimaryKey interface {
	PrimaryKey() []string
}

type WithUniqueIndexes interface {
	UniqueIndexes() Indexes
}

type WithIndexes interface {
	Indexes() Indexes
}

type WithComments interface {
	Comments() map[string]string
}

type WithRelations interface {
	ColRelations() map[string][]string
}

type WithColDescriptions interface {
	ColDescriptions() map[string][]string
}
