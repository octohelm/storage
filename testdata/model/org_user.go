package model

// OrgUser
// +gengo:table
// @def primary ID
// @def unique_index i_org_usr UserID OrgID
type OrgUser struct {
	ID     uint64 `db:"f_id,autoincrement"`
	UserID uint64 `db:"f_user_id"`
	OrgID  uint64 `db:"f_org_id"`
}
