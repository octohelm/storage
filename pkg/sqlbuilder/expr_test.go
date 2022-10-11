package sqlbuilder_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
	testingx "github.com/octohelm/x/testing"
)

func TestResolveExpr(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		testingx.Expect(t, ResolveExpr(nil), testingx.Be[*Ex](nil))
	})
}

type Byte uint8

func TestEx(t *testing.T) {
	t.Run("empty query", func(t *testing.T) {
		testutil.ShouldBeExpr(t, Expr(""), "")
	})

	t.Run("named arg", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Expr(`time > @left AND time < @right`, sql.Named("left", 1), sql.Named("right", 10)),
			"time > ? AND time < ?", 1, 10,
		)

		testutil.ShouldBeExpr(t,
			Expr(`time > @left AND time < @right`, NamedArgSet{
				"left":  1,
				"right": 10,
			}),
			"time > ? AND time < ?", 1, 10,
		)

		t.Run("nested named arg", func(t *testing.T) {
			testutil.ShouldBeExpr(t,
				Expr(`CREATE TABLE IF NOT EXISTS @table (@col)`, NamedArgSet{
					"table": Expr("t"),
					"col": Expr("@col @type", NamedArgSet{
						"col":  Expr("f_id"),
						"type": Expr("int"),
					}),
				}),
				"CREATE TABLE IF NOT EXISTS t (f_id int)",
			)
		})
	})

	t.Run("flatten slice", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Expr(`#ID IN (?)`, []int{28, 29, 30}),
			"#ID IN (?,?,?)", 28, 29, 30,
		)
	})

	t.Run("flatten slice for slice with named byte", func(t *testing.T) {
		fID := TypedCol[int]("f_id")
		fValue := TypedCol[Byte]("f_value")

		testutil.ShouldBeExpr(t,
			And(
				And(nil, fID.V(In(28))),
				fValue.V(In(Byte(28))),
			),
			"((f_id IN (?))) AND (f_value IN (?))", 28, Byte(28),
		)
	})

	t.Run("flatten should skip for bytes", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Expr(`#ID = (?)`, []byte("")),
			"#ID = (?)", []byte(""),
		)
	})

	t.Run("flatten with sub expr ", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Expr(`#ID = ?`, Expr("#ID + ?", 1)),
			"#ID = #ID + ?", 1,
		)
	})

	t.Run("flatten with ValuerExpr", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Expr(`#Point = ?`, Point{1, 1}),
			"#Point = ST_GeomFromText(?)", Point{1, 1},
		)
	})
}

func BenchmarkEx(b *testing.B) {
	b.Run("empty query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Expr("").Ex(context.Background())
		}
	})

	b.Run("flatten slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Expr(`#ID IN (?)`, []int{28, 29, 30}).Ex(context.Background())
		}
	})

	b.Run("flatten with sub expr", func(b *testing.B) {
		b.Run("raw", func(b *testing.B) {
			eb := Expr("")
			eb.Grow(2)

			eb.WriteQuery("#ID > ?")
			eb.WriteQuery(" AND ")
			eb.WriteQuery("#ID < ?")

			eb.AppendArgs(1, 10)

			rawBuild := func() *Ex {
				return eb.Ex(context.Background())
			}

			clone := func(ex *Ex) *Ex {
				return Expr(ex.Query(), ex.Args()...).Ex(context.Background())
			}

			b.Run("clone", func(b *testing.B) {
				ex := rawBuild()

				for i := 0; i < b.N; i++ {
					_ = clone(ex)
				}
			})
		})

		b.Run("IsNilExpr", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsNilExpr(Expr(`#ID > ?`, 1))
			}
		})

		b.Run("by expr", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e := And(
					TypedCol[int]("f_id").V(Lt(1)),
					TypedCol[int]("f_id").V(In(1, 2, 3)),
				)
				e.Ex(context.Background())
			}
		})

		b.Run("by expr without re created", func(b *testing.B) {
			fid := TypedCol[int]("f_id")
			left := fid.V(Lt(0))
			right := fid.V(In(1, 2, 3))

			b.Run("single", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					left.Ex(context.Background())
				}
			})

			b.Run("composed", func(b *testing.B) {
				e := And(left, left, right, right)

				b.Log(e.Ex(context.Background()).Query())

				for i := 0; i < b.N; i++ {
					e.Ex(context.Background())
				}
			})
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
