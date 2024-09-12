package model

import "github.com/octohelm/storage/pkg/sqlbuilder"

type UserV2 struct {
	ID       UserID `db:"f_id,autoincrement"`
	Nickname string `db:"f_nickname,size=255,default=''"`
	Gender   Gender `db:"f_gender,default='0'"`
	Name     string `db:"f_name,deprecated=f_real_name"`
	RealName string `db:"f_real_name,size=255,default=''"`
	// for modify testing
	Age      int8   `db:"f_age,default='0'"`
	Username string `db:"f_username,deprecated"`
}

func (user *UserV2) TableName() string {
	return "t_user"
}

func (user *UserV2) PrimaryKey() []string {
	return []string{"ID"}
}

func (user *UserV2) Indexes() sqlbuilder.Indexes {
	return sqlbuilder.Indexes{
		"i_nickname": {"Nickname"},
	}
}

func (user *UserV2) UniqueIndexes() sqlbuilder.Indexes {
	return sqlbuilder.Indexes{
		"i_name": {"RealName"},
		"i_age":  {"Age"},
	}
}
