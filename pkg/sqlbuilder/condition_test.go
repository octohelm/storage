package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestConditions(t *testing.T) {
	colA := TypedCol[int]("a")
	colB := TypedCol[string]("b")
	colC := TypedCol[int]("c")
	colD := TypedCol[string]("d")

	t.Run("Chain Condition", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				Or(
					And(
						nil,
						TypedCol[int]("a").V(Lt(1)),
						TypedCol[string]("b").V(LeftLike[string]("text")),
					),
					TypedCol[int]("a").V(Eq(2)),
				),
				TypedCol[string]("b").V(RightLike[string]("g")),
			),
			"(((a < ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text", 2, "g%",
		)
	})
	t.Run("Compose Condition", func(t *testing.T) {

		testutil.ShouldBeExpr(t,
			Xor(
				Or(
					And(
						(*Condition)(nil),
						(*Condition)(nil),
						(*Condition)(nil),
						(*Condition)(nil),
						colC.V(In(1, 2)),
						colC.V(In(3, 4)),
						colA.V(Eq(1)),
						colB.V(Like("text")),
					),
					colA.V(Eq(2)),
				),
				colB.V(Like("g")),
			),

			"(((c IN (?,?)) AND (c IN (?,?)) AND (a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, 2, 3, 4, 1, "%text%", 2, "%g%",
		)
	})
	t.Run("skip nil", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				colA.V(In[int]()),
				Or(
					colA.V(NotIn[int]()),
					And(
						nil,
						colA.V(Eq(1)),
						colB.V(Like("text")),
					),
					colA.V(Eq(2)),
				),
				colB.V(Like("g")),
			),
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		)
	})
	t.Run("XOR and OR", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				Or(
					colA.V(NotIn[int]()),
					And(
						nil,
						colA.V(Eq(1)),
						colB.V(Like("text")),
					),
					colA.V(Eq(2)),
				),
				colB.V(Like("g")),
			),
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		)
	})
	t.Run("XOR", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				colA.V(Eq(1)),
				colB.V(Like("g")),
			),
			"(a = ?) XOR (b LIKE ?)",
			1, "%g%",
		)
	})
	t.Run("Like", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(Like("e")),
			"d LIKE ?",
			"%e%",
		)
	})

	t.Run("Not like", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(NotLike("e")),
			"d NOT LIKE ?",
			"%e%",
		)
	})

	t.Run("Equal", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(Eq("e")),
			"d = ?", "e",
		)
	})
	t.Run("Not Equal", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(Neq("e")),
			"d <> ?",
			"e",
		)
	})
	t.Run("In", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(In("e", "f")),
			"d IN (?,?)", "e", "f",
		)
	})
	t.Run("NotIn", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(NotIn("e", "f")),
			"d NOT IN (?,?)", "e", "f",
		)
	})
	t.Run("Less than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(Lt(3)),
			"c < ?", 3,
		)
	})
	t.Run("Less or equal than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(Lte(3)),
			"c <= ?", 3,
		)
	})
	t.Run("Greater than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(Gt(3)),
			"c > ?", 3,
		)
	})
	t.Run("Greater or equal than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(Gte(3)),
			"c >= ?", 3,
		)
	})
	t.Run("Between", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(Between(0, 2)),
			"c BETWEEN ? AND ?", 0, 2,
		)
	})

	t.Run("Not between", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colC.V(NotBetween(0, 2)),
			"c NOT BETWEEN ? AND ?", 0, 2,
		)
	})

	t.Run("Is null", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(IsNull[string]()),
			"d IS NULL",
		)
	})

	t.Run("Is not null", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			colD.V(IsNotNull[string]()),
			"d IS NOT NULL",
		)
	})
}
