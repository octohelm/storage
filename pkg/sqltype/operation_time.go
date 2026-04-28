// Package sqltype 提供面向存储模型的通用字段类型与标记接口。
package sqltype

// WithCreationTime 表示模型支持标记创建时间。
type WithCreationTime interface {
	MarkCreatedAt()
}

// WithModificationTime 表示模型支持标记修改时间。
type WithModificationTime interface {
	MarkModifiedAt()
}
