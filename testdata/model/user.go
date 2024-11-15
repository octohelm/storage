package model

import "github.com/octohelm/storage/testdata/model/resource"

// User
// +gengo:table
// @def primary ID
// @def unique_index i_name Name DeletedAt
// @def unique_index i_age Age DeletedAt
// @def index i_nickname Nickname
// @def index i_created_at CreatedAt
type User struct {
	// 用户
	resource.Resource[UserID]

	// 姓名
	Name string `db:"f_name,size=255,default=''"`
	// 昵称
	Nickname string `db:"f_nickname,size=255,default=''"`
	// 用户名
	Username string `db:"f_username,default=''"`
	Gender   Gender `db:"f_gender,default='0'"`
	// 年龄
	Age int64 `db:"f_age,default='0'"`

	OperateTimeWithDeleted
}

type UserID uint64
