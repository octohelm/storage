package model

// User
// +gengo:table
// @def primary ID
// @def unique_index i_name Name DeletedAt
// @def unique_index i_age Age DeletedAt
// @def index i_nickname Nickname
// @def index i_created_at CreatedAt
type User struct {
	ID uint64 `db:"f_id,autoincrement"`
	// 姓名
	Name     string `db:"f_name,size=255,default=''"`
	Nickname string `db:"f_nickname,size=255,default=''"`
	Username string `db:"f_username,default=''"`
	Gender   Gender `db:"f_gender,default='0'"`
	Age      int64  `db:"f_age,default='0'"`
	OperateTimeWithDeleted
}
