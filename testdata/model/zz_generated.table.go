/*
Package model GENERATED BY gengo:table 
DON'T EDIT THIS FILE
*/
package model

import (
	context "context"

	github_com_octohelm_storage_pkg_sqlbuilder "github.com/octohelm/storage/pkg/sqlbuilder"
)

func (Org) TableName() string {
	return "t_org"
}

func (Org) Primary() []string {
	return []string{
		"ID",
	}
}

func (Org) Indexes() github_com_octohelm_storage_pkg_sqlbuilder.Indexes {
	return github_com_octohelm_storage_pkg_sqlbuilder.Indexes{
		"i_created_at": []string{
			"CreatedAt",
		},
	}
}

func (Org) UniqueIndexes() github_com_octohelm_storage_pkg_sqlbuilder.Indexes {
	return github_com_octohelm_storage_pkg_sqlbuilder.Indexes{
		"i_name": []string{
			"Name",
		},
	}
}

type tableOrg struct {
	ID        github_com_octohelm_storage_pkg_sqlbuilder.Column
	Name      github_com_octohelm_storage_pkg_sqlbuilder.Column
	CreatedAt github_com_octohelm_storage_pkg_sqlbuilder.Column
	UpdatedAt github_com_octohelm_storage_pkg_sqlbuilder.Column
	DeletedAt github_com_octohelm_storage_pkg_sqlbuilder.Column

	I     indexNameOfOrg
	table github_com_octohelm_storage_pkg_sqlbuilder.Table
}

func (tableOrg) New() github_com_octohelm_storage_pkg_sqlbuilder.Model {
	return &Org{}
}

func (t *tableOrg) IsNil() bool {
	return t.table.IsNil()
}

func (t *tableOrg) Ex(ctx context.Context) *github_com_octohelm_storage_pkg_sqlbuilder.Ex {
	return t.table.Ex(ctx)
}

func (t *tableOrg) TableName() string {
	return t.table.TableName()
}

func (t *tableOrg) F(name string) github_com_octohelm_storage_pkg_sqlbuilder.Column {
	return t.table.F(name)
}

func (t *tableOrg) K(name string) github_com_octohelm_storage_pkg_sqlbuilder.Key {
	return t.table.K(name)
}

func (t *tableOrg) Cols(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection {
	return t.table.Cols(names...)
}

func (t *tableOrg) Keys(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.KeyCollection {
	return t.table.Keys(names...)
}

type indexNameOfOrg struct {
	Primary github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
	IName   github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
}

var OrgT = &tableOrg{
	ID:        github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).F("ID"),
	Name:      github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).F("Name"),
	CreatedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).F("CreatedAt"),
	UpdatedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).F("UpdatedAt"),
	DeletedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).F("DeletedAt"),

	I: indexNameOfOrg{
		Primary: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).Cols([]string{
			"ID",
		}...),
		IName: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}).Cols([]string{
			"Name",
		}...),
	},
	table: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&Org{}),
}

func (OrgUser) TableName() string {
	return "t_org_user"
}

func (OrgUser) Primary() []string {
	return []string{
		"ID",
	}
}

func (OrgUser) UniqueIndexes() github_com_octohelm_storage_pkg_sqlbuilder.Indexes {
	return github_com_octohelm_storage_pkg_sqlbuilder.Indexes{
		"i_org_usr": []string{
			"UserID",
			"OrgID",
		},
	}
}

type tableOrgUser struct {
	ID     github_com_octohelm_storage_pkg_sqlbuilder.Column
	UserID github_com_octohelm_storage_pkg_sqlbuilder.Column
	OrgID  github_com_octohelm_storage_pkg_sqlbuilder.Column

	I     indexNameOfOrgUser
	table github_com_octohelm_storage_pkg_sqlbuilder.Table
}

func (tableOrgUser) New() github_com_octohelm_storage_pkg_sqlbuilder.Model {
	return &OrgUser{}
}

func (t *tableOrgUser) IsNil() bool {
	return t.table.IsNil()
}

func (t *tableOrgUser) Ex(ctx context.Context) *github_com_octohelm_storage_pkg_sqlbuilder.Ex {
	return t.table.Ex(ctx)
}

func (t *tableOrgUser) TableName() string {
	return t.table.TableName()
}

func (t *tableOrgUser) F(name string) github_com_octohelm_storage_pkg_sqlbuilder.Column {
	return t.table.F(name)
}

func (t *tableOrgUser) K(name string) github_com_octohelm_storage_pkg_sqlbuilder.Key {
	return t.table.K(name)
}

func (t *tableOrgUser) Cols(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection {
	return t.table.Cols(names...)
}

func (t *tableOrgUser) Keys(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.KeyCollection {
	return t.table.Keys(names...)
}

type indexNameOfOrgUser struct {
	Primary github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
	IOrgUsr github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
}

var OrgUserT = &tableOrgUser{
	ID:     github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}).F("ID"),
	UserID: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}).F("UserID"),
	OrgID:  github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}).F("OrgID"),

	I: indexNameOfOrgUser{
		Primary: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}).Cols([]string{
			"ID",
		}...),
		IOrgUsr: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}).Cols([]string{
			"UserID",
			"OrgID",
		}...),
	},
	table: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&OrgUser{}),
}

func (User) TableName() string {
	return "t_user"
}

func (User) Primary() []string {
	return []string{
		"ID",
	}
}

func (User) Indexes() github_com_octohelm_storage_pkg_sqlbuilder.Indexes {
	return github_com_octohelm_storage_pkg_sqlbuilder.Indexes{
		"i_created_at": []string{
			"CreatedAt",
		},
		"i_nickname": []string{
			"Nickname",
		},
	}
}

func (User) UniqueIndexes() github_com_octohelm_storage_pkg_sqlbuilder.Indexes {
	return github_com_octohelm_storage_pkg_sqlbuilder.Indexes{
		"i_age": []string{
			"Age",
			"DeletedAt",
		},
		"i_name": []string{
			"Name",
			"DeletedAt",
		},
	}
}

type tableUser struct {
	ID        github_com_octohelm_storage_pkg_sqlbuilder.Column
	Name      github_com_octohelm_storage_pkg_sqlbuilder.Column
	Nickname  github_com_octohelm_storage_pkg_sqlbuilder.Column
	Username  github_com_octohelm_storage_pkg_sqlbuilder.Column
	Gender    github_com_octohelm_storage_pkg_sqlbuilder.Column
	Age       github_com_octohelm_storage_pkg_sqlbuilder.Column
	CreatedAt github_com_octohelm_storage_pkg_sqlbuilder.Column
	UpdatedAt github_com_octohelm_storage_pkg_sqlbuilder.Column
	DeletedAt github_com_octohelm_storage_pkg_sqlbuilder.Column

	I     indexNameOfUser
	table github_com_octohelm_storage_pkg_sqlbuilder.Table
}

func (tableUser) New() github_com_octohelm_storage_pkg_sqlbuilder.Model {
	return &User{}
}

func (t *tableUser) IsNil() bool {
	return t.table.IsNil()
}

func (t *tableUser) Ex(ctx context.Context) *github_com_octohelm_storage_pkg_sqlbuilder.Ex {
	return t.table.Ex(ctx)
}

func (t *tableUser) TableName() string {
	return t.table.TableName()
}

func (t *tableUser) F(name string) github_com_octohelm_storage_pkg_sqlbuilder.Column {
	return t.table.F(name)
}

func (t *tableUser) K(name string) github_com_octohelm_storage_pkg_sqlbuilder.Key {
	return t.table.K(name)
}

func (t *tableUser) Cols(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection {
	return t.table.Cols(names...)
}

func (t *tableUser) Keys(names ...string) github_com_octohelm_storage_pkg_sqlbuilder.KeyCollection {
	return t.table.Keys(names...)
}

type indexNameOfUser struct {
	Primary github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
	IName   github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
	IAge    github_com_octohelm_storage_pkg_sqlbuilder.ColumnCollection
}

var UserT = &tableUser{
	ID:        github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("ID"),
	Name:      github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("Name"),
	Nickname:  github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("Nickname"),
	Username:  github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("Username"),
	Gender:    github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("Gender"),
	Age:       github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("Age"),
	CreatedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("CreatedAt"),
	UpdatedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("UpdatedAt"),
	DeletedAt: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).F("DeletedAt"),

	I: indexNameOfUser{
		Primary: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).Cols([]string{
			"ID",
		}...),
		IName: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).Cols([]string{
			"Name",
			"DeletedAt",
		}...),
		IAge: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}).Cols([]string{
			"Age",
			"DeletedAt",
		}...),
	},
	table: github_com_octohelm_storage_pkg_sqlbuilder.TableFromModel(&User{}),
}
