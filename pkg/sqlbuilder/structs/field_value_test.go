package structs

import (
	"context"
	"reflect"
	"slices"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

type auditOnly struct {
	Name string `db:"f_name"`
	Old  string `db:"f_old,deprecated"`
}

type userModel struct {
	Name string `db:"f_name"`
}

func (userModel) TableName() string { return "t_user" }

type orgModel struct {
	Name string `db:"f_name"`
}

func (orgModel) TableName() string { return "t_org" }

type aliasedScalar struct {
	Name string `db:"f_name" alias:"t_org"`
}

func TestAllFieldValue(t *testing.T) {
	values := slices.Collect(AllFieldValue(context.Background(), &auditOnly{
		Name: "alice",
		Old:  "legacy",
	}))

	Then(
		t, "AllFieldValue 会跳过 deprecated 字段",
		Expect(len(values), Equal(1)),
		Expect(values[0].Field.FieldName, Equal("Name")),
		Expect(values[0].Value.Interface(), Equal(any("alice"))),
	)

	modelValues := slices.Collect(AllFieldValue(context.Background(), &userModel{Name: "alice"}))
	Then(
		t, "AllFieldValue 会继承模型表名",
		Expect(modelValues[0].TableName, Equal("t_user")),
	)

	aliased := slices.Collect(AllFieldValue(context.Background(), &aliasedScalar{Name: "acme"}))

	orgName := slices.IndexFunc(aliased, func(v *FieldValue) bool {
		return v.Field.FieldName == "Name" && v.TableName == "t_org"
	})

	Then(
		t, "AllFieldValue 会尊重 alias 覆盖",
		Expect(orgName >= 0, Equal(true)),
	)
}

func TestAllFieldValueOmitZero(t *testing.T) {
	var ptr *auditOnly
	rv := reflect.ValueOf(&ptr).Elem()

	omitted := slices.Collect(AllFieldValueOmitZero(context.Background(), rv, "Name"))

	Then(
		t, "AllFieldValueOmitZero 会初始化空指针 reflect.Value 并保留显式排除字段",
		Expect(ptr != nil, Equal(true)),
		Expect(len(omitted), Equal(1)),
		Expect(omitted[0].Field.FieldName, Equal("Name")),
	)
}
