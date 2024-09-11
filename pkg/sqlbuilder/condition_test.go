package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"
	testingx "github.com/octohelm/x/testing"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestConditions(t *testing.T) {
	colA := TypedCol[int]("a")
	colB := TypedCol[string]("b")
	colC := TypedCol[int]("c")
	colD := TypedCol[string]("d")

	t.Run("Chain Condition", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
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
			testutil.BeFragment(
				"(((a < ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%text", 2, "g%",
			),
		)
	})

	t.Run("Compose Condition", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
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

			testutil.BeFragment("(((c IN (?,?)) AND (c IN (?,?)) AND (a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, 2, 3, 4, 1, "%text%", 2, "%g%",
			),
		)
	})
	t.Run("skip nil", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
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
			testutil.BeFragment("(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%text%", 2, "%g%",
			),
		)
	})
	t.Run("XOR and OR", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
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
			testutil.BeFragment("(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
				1, "%text%", 2, "%g%",
			),
		)
	})
	t.Run("XOR", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			Xor(
				colA.V(Eq(1)),
				colB.V(Like("g")),
			),
			testutil.BeFragment("(a = ?) XOR (b LIKE ?)",
				1, "%g%",
			),
		)
	})
	t.Run("Like", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(Like("e")),
			testutil.BeFragment("d LIKE ?",
				"%e%",
			))
	})

	t.Run("Not like", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(NotLike("e")),
			testutil.BeFragment("d NOT LIKE ?",
				"%e%",
			),
		)
	})

	t.Run("Equal", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(Eq("e")),
			testutil.BeFragment("d = ?", "e"),
		)
	})
	t.Run("Not Equal", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(Neq("e")),
			testutil.BeFragment("d <> ?",
				"e",
			),
		)
	})
	t.Run("In", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(In("e", "f")),
			testutil.BeFragment("d IN (?,?)", "e", "f"),
		)
	})
	t.Run("NotIn", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(NotIn("e", "f")),
			testutil.BeFragment("d NOT IN (?,?)", "e", "f"),
		)
	})
	t.Run("Less than", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(Lt(3)),
			testutil.BeFragment("c < ?", 3),
		)
	})
	t.Run("Less or equal than", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(Lte(3)),
			testutil.BeFragment("c <= ?", 3),
		)
	})
	t.Run("Greater than", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(Gt(3)),
			testutil.BeFragment("c > ?", 3),
		)
	})
	t.Run("Greater or equal than", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(Gte(3)),
			testutil.BeFragment("c >= ?", 3),
		)
	})
	t.Run("Between", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(Between(0, 2)),
			testutil.BeFragment("c BETWEEN ? AND ?", 0, 2),
		)
	})

	t.Run("Not between", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colC.V(NotBetween(0, 2)),
			testutil.BeFragment("c NOT BETWEEN ? AND ?", 0, 2),
		)
	})

	t.Run("Is null", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(IsNull[string]()),
			testutil.BeFragment("d IS NULL"),
		)
	})

	t.Run("Is not null", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			colD.V(IsNotNull[string]()),
			testutil.BeFragment("d IS NOT NULL"),
		)
	})
}
