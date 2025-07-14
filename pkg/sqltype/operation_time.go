package sqltype

type WithCreationTime interface {
	MarkCreatedAt()
}

type WithModificationTime interface {
	MarkModifiedAt()
}
