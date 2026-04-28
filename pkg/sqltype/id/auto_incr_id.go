package id

// AutoIncrID 定义通用的自增主键字段。
type AutoIncrID struct {
	// 自增，保留字段
	ID uint64 `db:"f_id,autoincrement" json:"-"`
}
