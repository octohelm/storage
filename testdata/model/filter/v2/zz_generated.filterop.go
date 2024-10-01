/*
Package filter GENERATED BY gengo:filterop 
DON'T EDIT THIS FILE
*/
package filter

import (
	datatypes "github.com/octohelm/storage/pkg/datatypes"
	filter "github.com/octohelm/storage/pkg/filter"
	sqlpipe "github.com/octohelm/storage/pkg/sqlpipe"
	sqlpipefilter "github.com/octohelm/storage/pkg/sqlpipe/filter"
	model "github.com/octohelm/storage/testdata/model"
)

type UserByID struct {
	// 按  筛选
	ID *filter.Filter[model.UserID] `name:"user~id,omitempty" in:"query"`
}

func (f *UserByID) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByID) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.ID, f.ID))
}

type UserByName struct {
	// 按 姓名 筛选
	Name *filter.Filter[string] `name:"user~name,omitempty" in:"query"`
}

func (f *UserByName) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByName) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.Name, f.Name))
}

type UserByNickname struct {
	// 按  筛选
	Nickname *filter.Filter[string] `name:"user~nickname,omitempty" in:"query"`
}

func (f *UserByNickname) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByNickname) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.Nickname, f.Nickname))
}

type UserByAge struct {
	// 按  筛选
	Age *filter.Filter[int64] `name:"user~age,omitempty" in:"query"`
}

func (f *UserByAge) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByAge) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.Age, f.Age))
}

type UserByCreatedAt struct {
	// 按  筛选
	CreatedAt *filter.Filter[datatypes.Datetime] `name:"user~createdAt,omitempty" in:"query"`
}

func (f *UserByCreatedAt) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByCreatedAt) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.CreatedAt, f.CreatedAt))
}

type UserByDeletedAt struct {
	// 按  筛选
	DeletedAt *filter.Filter[int64] `name:"user~deletedAt,omitempty" in:"query"`
}

func (f *UserByDeletedAt) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *UserByDeletedAt) Next(src sqlpipe.Source[model.User]) sqlpipe.Source[model.User] {
	return src.Pipe(sqlpipefilter.AsWhere(model.UserT.DeletedAt, f.DeletedAt))
}

type OrgByID struct {
	// 按  筛选
	ID *filter.Filter[model.OrgID] `name:"org~id,omitempty" in:"query"`
}

func (f *OrgByID) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgByID) Next(src sqlpipe.Source[model.Org]) sqlpipe.Source[model.Org] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgT.ID, f.ID))
}

type OrgByName struct {
	// 按  筛选
	Name *filter.Filter[string] `name:"org~name,omitempty" in:"query"`
}

func (f *OrgByName) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgByName) Next(src sqlpipe.Source[model.Org]) sqlpipe.Source[model.Org] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgT.Name, f.Name))
}

type OrgByCreatedAt struct {
	// 按  筛选
	CreatedAt *filter.Filter[datatypes.Datetime] `name:"org~createdAt,omitempty" in:"query"`
}

func (f *OrgByCreatedAt) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgByCreatedAt) Next(src sqlpipe.Source[model.Org]) sqlpipe.Source[model.Org] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgT.CreatedAt, f.CreatedAt))
}

type OrgUserByID struct {
	// 按  筛选
	ID *filter.Filter[uint64] `name:"org-user~id,omitempty" in:"query"`
}

func (f *OrgUserByID) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgUserByID) Next(src sqlpipe.Source[model.OrgUser]) sqlpipe.Source[model.OrgUser] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgUserT.ID, f.ID))
}

type OrgUserByUserID struct {
	// 按  筛选
	UserID *filter.Filter[model.UserID] `name:"org-user~userId,omitempty" in:"query"`
}

func (f *OrgUserByUserID) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgUserByUserID) Next(src sqlpipe.Source[model.OrgUser]) sqlpipe.Source[model.OrgUser] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgUserT.UserID, f.UserID))
}

type OrgUserByOrgID struct {
	// 按  筛选
	OrgID *filter.Filter[model.OrgID] `name:"org-user~orgId,omitempty" in:"query"`
}

func (f *OrgUserByOrgID) OperatorType() sqlpipe.OperatorType {
	return sqlpipe.OperatorFilter
}

func (f *OrgUserByOrgID) Next(src sqlpipe.Source[model.OrgUser]) sqlpipe.Source[model.OrgUser] {
	return src.Pipe(sqlpipefilter.AsWhere(model.OrgUserT.OrgID, f.OrgID))
}