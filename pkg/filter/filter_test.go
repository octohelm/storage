package filter

import (
	"encoding"
	"reflect"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestFilterMarshalAndUnmarshal(t *testing.T) {
	cases := []struct {
		filter Rule
		query  string
	}{
		{
			Eq(1),
			"eq(1)",
		},
		{
			And[int](Gte(1), Lte(10)),
			"and(gte(1),lte(10))",
		},
		{
			And[int](Gte(1), Lte(10)).WhereOf("item.id"),
			`where("item.id",and(gte(1),lte(10)))`,
		},
		{
			In([]string{"a", "b", "c"}).WhereOf("item.name"),
			`where("item.name",in("a","b","c"))`,
		},
	}

	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			data, err := c.filter.MarshalText()
			testingx.Expect(t, err, testingx.BeNil[error]())
			testingx.Expect(t, string(data), testingx.Be(c.query))

			tt := reflect.TypeOf(c.filter)
			if tt.Kind() == reflect.Ptr {
				tt = tt.Elem()
			}

			f := reflect.New(tt).Interface().(Rule)

			err = f.(encoding.TextUnmarshaler).UnmarshalText([]byte(c.query))
			testingx.Expect(t, err, testingx.BeNil[error]())

			raw, err := f.MarshalText()
			testingx.Expect(t, err, testingx.BeNil[error]())
			testingx.Expect(t, string(raw), testingx.Equal(string(data)))
		})
	}

	t.Run("unmarshal single value", func(t *testing.T) {
		f := Filter[int]{}
		err := f.UnmarshalText([]byte("1"))
		testingx.Expect(t, err, testingx.BeNil[error]())
		testingx.Expect(t, f.String(), testingx.Be("eq(1)"))
	})

	t.Run("composed", func(t *testing.T) {
		c := Compose(
			ItemListFilterByName{
				ItemName: Eq("x"),
			},
			ItemListFilterByID{
				ItemID: In([]int{1, 2, 3, 4}),
			},
		)

		rule, err := c.MarshalText()
		testingx.Expect(t, err, testingx.BeNil[error]())
		testingx.Expect(t, string(rule), testingx.Be(`or(where("item.name",eq("x")),where("item.id",in(1,2,3,4)))`))

		cc := Compose(
			ItemListFilterByName{},
			ItemListFilterByID{},
		)

		err = cc.UnmarshalText(rule)
		testingx.Expect(t, err, testingx.BeNil[error]())

		rule2, err := cc.MarshalText()
		testingx.Expect(t, err, testingx.BeNil[error]())
		testingx.Expect(t, string(rule2), testingx.Be(`or(where("item.name",eq("x")),where("item.id",in(1,2,3,4)))`))
	})
}

type ItemListFilterByID struct {
	ItemID *Filter[int] `name:"item.id,omitempty" in:"query"`
}

type ItemListFilterByName struct {
	ItemName *Filter[string] `name:"item.name,omitempty" in:"query"`
}
