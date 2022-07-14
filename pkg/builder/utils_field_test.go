package builder_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/octohelm/sqlx/internal/testutil"

	. "github.com/octohelm/sqlx/pkg/builder"
	"github.com/octohelm/x/ptr"
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
	fields := StructFieldsFor(context.Background(), typex.FromRType(reflect.TypeOf(p)))

	rv := reflect.ValueOf(p)

	testutil.Expect(t, fields, testutil.HaveLen[[]*StructField](5))

	testutil.Expect(t, fields[0].Name, testutil.Equal("f_x"))
	testutil.Expect(t, fields[0].FieldValue(rv).Interface(), testutil.Equal[any](p.X))

	testutil.Expect(t, fields[1].Name, testutil.Equal("f_a"))
	testutil.Expect(t, fields[1].FieldValue(rv).Interface(), testutil.Equal[any](p.A))

	testutil.Expect(t, fields[2].Name, testutil.Equal("f_b"))
	testutil.Expect(t, fields[2].FieldValue(rv).Interface(), testutil.Equal[any](p.B))

	testutil.Expect(t, fields[3].Name, testutil.Equal("f_b_map"))
	testutil.Expect(t, fields[3].FieldValue(rv).Interface(), testutil.Equal[any](p.Map))

	testutil.Expect(t, fields[4].Name, testutil.Equal("f_c"))
	testutil.Expect(t, fields[4].FieldValue(rv).Interface(), testutil.Equal[any](p.C))
}

func BenchmarkTableFieldsFor(b *testing.B) {
	typeP := reflect.TypeOf(p)

	_ = StructFieldsFor(context.Background(), typex.FromRType(typeP))

	//b.Log(typex.FromRType(reflect.TypeOf(p)).Unwrap() == typex.FromRType(reflect.TypeOf(p)).Unwrap())

	b.Run("StructFieldsFor", func(b *testing.B) {
		typP := typex.FromRType(typeP)

		for i := 0; i < b.N; i++ {
			_ = StructFieldsFor(context.Background(), typP)
		}
	})
}
