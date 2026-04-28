package internal

// Model 表示可映射到表名的模型。
type Model interface {
	TableName() string
}

// ModelNewer 表示可分配模型实例的工厂约束。
type ModelNewer[M Model] interface {
	Model() *M
}
