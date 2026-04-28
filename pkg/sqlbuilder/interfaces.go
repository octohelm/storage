package sqlbuilder

import (
	"github.com/octohelm/storage/pkg/sqlbuilder/internal"
)

// Model 复用内部模型约束。
type Model = internal.Model

// ModelNewer 复用内部模型构造约束。
type ModelNewer[M Model] internal.ModelNewer[M]

// DataTypeDescriber 表示类型可按驱动名描述数据库字段类型。
type DataTypeDescriber interface {
	DataType(driverName string) string
}

// WithTableDescription 表示模型提供表级说明。
type WithTableDescription interface {
	TableDescription() []string
}

// Indexes 表示索引名到字段列表的映射。
type Indexes map[string][]string

// WithPrimaryKey 表示模型声明主键字段。
type WithPrimaryKey interface {
	PrimaryKey() []string
}

// WithUniqueIndexes 表示模型声明唯一索引。
type WithUniqueIndexes interface {
	UniqueIndexes() Indexes
}

// WithIndexes 表示模型声明普通索引。
type WithIndexes interface {
	Indexes() Indexes
}

// WithComments 表示模型声明字段注释。
type WithComments interface {
	Comments() map[string]string
}

// WithRelations 表示模型声明字段关联关系。
type WithRelations interface {
	ColRelations() map[string][]string
}

// WithColDescriptions 表示模型声明字段说明。
type WithColDescriptions interface {
	ColDescriptions() map[string][]string
}
