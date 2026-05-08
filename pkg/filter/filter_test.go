package filter

import (
	"encoding"
	"reflect"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestFilter(t *testing.T) {
	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		cases := []struct {
			filter Rule
			query  string
		}{
			{Eq(1), "eq(1)"},
			{And[int](Gte(1), Lte(10)), "and(gte(1),lte(10))"},
			{And[int](Gte(1), Lte(10)).WhereOf("item.id"), `where("item.id",and(gte(1),lte(10)))`},
			{In("a", "b", "c").WhereOf("item.name"), `where("item.name",in("a","b","c"))`},
		}

		for _, c := range cases {
			t.Run(c.query, func(t *testing.T) {
				data := MustValue(t, c.filter.MarshalText)

				Then(
					t, "Marshal 结果应符合预期",
					Expect(string(data), Equal(c.query)),
				)

				tt := reflect.TypeOf(c.filter)
				if tt.Kind() == reflect.Pointer {
					tt = tt.Elem()
				}
				f := reflect.New(tt).Interface().(Rule)

				Must(t, func() error {
					return f.(encoding.TextUnmarshaler).UnmarshalText(data)
				})

				Then(
					t, "Unmarshal 且再次 Marshal 后结果应一致",
					ExpectMustValue(f.MarshalText, Equal(data)),
				)
			})
		}
	})

	t.Run("Unmarshal single value", func(t *testing.T) {
		f := Filter[int]{}

		Must(t, func() error {
			return f.UnmarshalText([]byte("1"))
		})

		Then(
			t, "单值应解析为 eq 规则",
			Expect(f.String(), Equal("eq(1)")),
		)
	})

	t.Run("Composed Filters", func(t *testing.T) {
		c := Compose(
			ItemListFilterByName{ItemName: Eq("x")},
			ItemListFilterByID{ItemID: In(1, 2, 3, 4)},
		)

		expected := `or(where("item.name",eq("x")),where("item.id",in(1,2,3,4)))`

		Then(
			t, "组合 Filter 应正确序列化",
			ExpectMustValue(c.MarshalText, Equal([]byte(expected))),
		)

		t.Run("Round trip", func(t *testing.T) {
			cc := Compose(ItemListFilterByName{}, ItemListFilterByID{})

			Must(t, func() error {
				return cc.UnmarshalText([]byte(expected))
			})

			Then(
				t, "反序列化后应能还原相同的规则",
				ExpectMustValue(cc.MarshalText, Equal([]byte(expected))),
			)
		})
	})

	t.Run("Boundary Conditions", func(t *testing.T) {
		t.Run("Nil filter handling", func(t *testing.T) {
			var f *Filter[int]

			Then(
				t, "空指针应被识别为 Nil",
				Expect(f, Be(cmp.Nil[*Filter[int]]())),
			)
		})
	})
}

type ItemListFilterByID struct {
	ItemID *Filter[int] `name:"item.id,omitzero" in:"query"`
}

type ItemListFilterByName struct {
	ItemName *Filter[string] `name:"item.name,omitzero" in:"query"`
}
