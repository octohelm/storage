package sqlfrag_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"slices"
	"testing"

	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
)

func TestFragment(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Const(""),
			testutil.BeFragment(""),
		)
	})

	t.Run("const fragment", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Const("SELECT 1"),
			testutil.BeFragment("SELECT 1"),
		)
	})

	t.Run("flatten seq", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Pair(`#ID IN (?)`, slices.Values([]any{28, 29, 30})),
			testutil.BeFragment("#ID IN (?,?,?)", 28, 29, 30),
		)
	})

	t.Run("flatten typed seq", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Pair(`#ID IN (?)`, slices.Values([]int{28, 29, 30})),
			testutil.BeFragment("#ID IN (?,?,?)", 28, 29, 30),
		)
	})

	t.Run("flatten slice", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Pair(`#ID IN (?)`, []int{28, 29, 30}),
			testutil.BeFragment("#ID IN (?,?,?)", 28, 29, 30),
		)
	})

	t.Run("flatten slice composed", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Pair(`DO UPDATE SET f_name = ?`, []any{
				sqlfrag.Pair("EXCLUDED.?", sqlfrag.Const("f_name")),
			}),
			testutil.BeFragment("DO UPDATE SET f_name = EXCLUDED.f_name"),
		)
	})

	t.Run("flatten with sub frag ", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			sqlfrag.Pair(`#ID = ?`, sqlfrag.Pair("#ID + ?", 1)),
			testutil.BeFragment("#ID = #ID + ?", 1),
		)
	})

	t.Run("flatten with CustomValueArg", func(t *testing.T) {
		testingx.Expect(t,
			sqlfrag.Pair(`#Point = ?`, Point{1, 1}),
			testutil.BeFragment("#Point = ST_GeomFromText(?)", Point{1, 1}),
		)
	})

	t.Run("named arg", func(t *testing.T) {
		t.Run("with named arg", func(t *testing.T) {
			testingx.Expect(t,
				sqlfrag.Pair(`time > @left AND time < @right`, sql.Named("left", 1), sql.Named("right", 10)),
				testutil.BeFragment("time > ? AND time < ?", 1, 10),
			)
		})

		t.Run("with named arg set", func(t *testing.T) {
			testingx.Expect(t,
				sqlfrag.Pair(`time > @left AND time < @right`, sqlfrag.NamedArgSet{
					"left":  1,
					"right": 10,
				}),
				testutil.BeFragment("time > ? AND time < ?", 1, 10),
			)
		})

		t.Run("deep nested", func(t *testing.T) {
			testingx.Expect(t,
				sqlfrag.Pair(`CREATE TABLE IF NOT EXISTS @table @col`, sqlfrag.NamedArgSet{
					"table": sqlfrag.Pair("t"),
					"col": sqlfrag.Block(
						sqlfrag.Pair("\n@col @type", sqlfrag.NamedArgSet{
							"col":  sqlfrag.Pair("f_id"),
							"type": sqlfrag.Pair("int"),
						}),
					),
				}),
				testutil.BeFragment(`
CREATE TABLE IF NOT EXISTS t (
	f_id int
)
`),
			)
		})
	})
}

type Point struct {
	X float64
	Y float64
}

func (Point) DataType(engine string) string {
	return "POINT"
}

func (Point) ValueEx() string {
	return `ST_GeomFromText(?)`
}

func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%v %v)", p.X, p.Y), nil
}
