package internal

type Model interface {
	TableName() string
}

type ModelNewer[M Model] interface {
	Model() *M
}
