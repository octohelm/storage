package id

type AutoIncrID struct {
	// 自增，保留字段
	ID uint64 `db:"f_id,autoincrement" json:"-"`
}
