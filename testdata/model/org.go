package model

// Org
// +gengo:table
// @def primary ID
// @def unique_index i_name Name
// @def index i_created_at CreatedAt
type Org struct {
	ID   uint64 `db:"f_id,autoincrement"`
	Name string `db:"f_name,size=255,default=''"`
	OperateTimeWithDeleted
}
