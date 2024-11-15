package resource

type Resource[ID ~uint64] struct {
	// ID
	// 生成 ID
	ID ID `db:"f_id,autoincrement"`
}
