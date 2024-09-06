package structs

import (
	"context"
	"reflect"
	"testing"

	"github.com/octohelm/x/ptr"
	testingx "github.com/octohelm/x/testing"
	typex "github.com/octohelm/x/types"
)

type SubSubSub struct {
	X string `db:"f_x"`
}

type SubSub struct {
	SubSubSub
}

type Sub struct {
	SubSub
	A string `db:"f_a"`
}

type PtrSub struct {
	B   []string          `db:"f_b"`
	Map map[string]string `db:"f_b_map"`
}

type P struct {
	Sub
	*PtrSub
	C *string `db:"f_c"`
}

var p *P

func init() {
	p = &P{}
	p.X = "x"
	p.A = "a"
	p.PtrSub = &PtrSub{
		Map: map[string]string{
			"1": "!",
		},
		B: []string{"b"},
	}
	p.C = ptr.String("c")
}

func TestTableFieldsFor(t *testing.T) {
	fields := Fields(context.Background(), typex.FromRType(reflect.TypeOf(p)))

	rv := reflect.ValueOf(p)

	testingx.Expect(t, fields, testingx.HaveLen[[]*Field](5))

	testingx.Expect(t, fields[0].Name, testingx.Equal("f_x"))
	testingx.Expect(t, fields[0].FieldValue(rv).Interface(), testingx.Equal[any](p.X))

	testingx.Expect(t, fields[1].Name, testingx.Equal("f_a"))
	testingx.Expect(t, fields[1].FieldValue(rv).Interface(), testingx.Equal[any](p.A))

	testingx.Expect(t, fields[2].Name, testingx.Equal("f_b"))
	testingx.Expect(t, fields[2].FieldValue(rv).Interface(), testingx.Equal[any](p.B))

	testingx.Expect(t, fields[3].Name, testingx.Equal("f_b_map"))
	testingx.Expect(t, fields[3].FieldValue(rv).Interface(), testingx.Equal[any](p.Map))

	testingx.Expect(t, fields[4].Name, testingx.Equal("f_c"))
	testingx.Expect(t, fields[4].FieldValue(rv).Interface(), testingx.Equal[any](p.C))
}

func BenchmarkTableFieldsFor(b *testing.B) {
	typeP := reflect.TypeOf(p)

	_ = Fields(context.Background(), typex.FromRType(typeP))

	//b.Log(typex.FromRType(reflect.TypeOf(p)).Unwrap() == typex.FromRType(reflect.TypeOf(p)).Unwrap())

	b.Run("StructFieldsFor", func(b *testing.B) {
		typP := typex.FromRType(typeP)

		for i := 0; i < b.N; i++ {
			_ = Fields(context.Background(), typP)
		}
	})
}

//
//func TestValueMap(t *testing.T) {
//	type User struct {
//		ID       uint64 `db:"f_id"`
//		Name     string `db:"f_name"`
//		Username string `db:"f_username"`
//	}
//
//	user := User{
//		ID: 123123213,
//	}
//
//	t.Run("#FieldValuesFromStructBy", func(t *testing.T) {
//		testingx.Expect(t, FieldValuesFromStructBy(user, []string{}), testingx.HaveLen[FieldValues](0))
//
//		values := FieldValuesFromStructBy(user, []string{"ID"})
//
//		testingx.Expect(t, values, testingx.Equal(FieldValues{
//			"ID": user.ID,
//		}))
//	})
//
//	t.Run("#FieldValuesFromStructBy", func(t *testing.T) {
//		testingx.Expect(t, FieldValuesFromStructByNonZero(user), testingx.Equal(FieldValues{
//			"ID": user.ID,
//		}))
//
//		testingx.Expect(t, FieldValuesFromStructByNonZero(user, "Username"), testingx.Equal(FieldValues{
//			"ID":       user.ID,
//			"Username": user.Username,
//		}))
//	})
//
//	t.Run("#GetColumnName", func(t *testing.T) {
//		testingx.Expect(t, GetColumnName("Text", ""), testingx.Equal("f_text"))
//		testingx.Expect(t, GetColumnName("Text", ",size=256"), testingx.Equal("f_text"))
//		testingx.Expect(t, GetColumnName("Text", "f_text2"), testingx.Equal("f_text2"))
//		testingx.Expect(t, GetColumnName("Text", "f_text2,default=''"), testingx.Equal("f_text2"))
//	})
//}
//
//func TestParseDef(t *testing.T) {
//	t.Run("index with Field Names", func(t *testing.T) {
//
//		i := ParseIndexDefine("index i_xxx/BTREE Name")
//
//		testingx.Expect(t, i, testingx.Equal(&IndexDefine{
//			Kind:              "index",
//			Name:              "i_xxx",
//			Method:            "BTREE",
//			ColNameAndOptions: []string{"Name"},
//		}))
//	})
//
//	t.Run("primary with Field Names", func(t *testing.T) {
//
//		i := ParseIndexDefine("primary ID Name")
//
//		testingx.Expect(t, i, testingx.Equal(&IndexDefine{
//			Kind:              "primary",
//			ColNameAndOptions: []string{"ID", "Name"},
//		}))
//	})
//
//	t.Run("index with expr", func(t *testing.T) {
//		i := ParseIndexDefine("index i_xxx/GIST Test/gist_trgm_ops")
//
//		testingx.Expect(t, i, testingx.Equal(&IndexDefine{
//			Kind:   "index",
//			Name:   "i_xxx",
//			Method: "GIST",
//			ColNameAndOptions: []string{
//				"Test/gist_trgm_ops",
//			},
//		}))
//	})
//}
//
//type UserBase struct {
//	ID       uint64 `db:"f_id"`
//	Name     string `db:"f_name"`
//	Username string `db:"f_username"`
//}
//
//func (UserBase) TableName() string {
//	return "t_user_base"
//}
//
//type User struct {
//	UserBase
//}
//
//func (User) TableName() string {
//	return "t_user"
//}
//
//type OrgUser struct {
//	OrgID  uint64 `db:"f_org_id"`
//	UserID uint64 `db:"f_user_id"`
//}
//
//func (OrgUser) TableName() string {
//	return "t_org_user"
//}
//
//type Org struct {
//	ID   uint64 `db:"f_id"`
//	Name string `db:"f_name"`
//}
//
//func (Org) TableName() string {
//	return "t_org"
//}
//
//type OrgUserAll struct {
//	OrgUser
//	User User `json:"user"`
//	Org  Org  `json:"org"`
//}
//
//func TestColumnsByStruct(t *testing.T) {
//	t.Run("simple", func(t *testing.T) {
//		q := ColumnsByStruct(&User{}).Ex(context.Background()).Query()
//		testingx.Expect(t, q, testingx.Equal("t_user.f_id AS t_user__f_id, t_user.f_name AS t_user__f_name, t_user.f_username AS t_user__f_username"))
//	})
//
//	t.Run("joined", func(t *testing.T) {
//		q := ColumnsByStruct(&OrgUserAll{}).Ex(context.Background()).Query()
//		for _, g := range strings.Split(q, ", ") {
//			t.Log(g)
//		}
//		testingx.Expect(t, q, testingx.Equal("t_org_user.f_org_id AS t_org_user__f_org_id, t_org_user.f_user_id AS t_org_user__f_user_id, t_user.f_id AS t_user__f_id, t_user.f_name AS t_user__f_name, t_user.f_username AS t_user__f_username, t_org.f_id AS t_org__f_id, t_org.f_name AS t_org__f_name"))
//	})
//}
