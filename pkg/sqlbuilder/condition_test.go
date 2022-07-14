package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/internal/testutil"
	. "github.com/octohelm/storage/pkg/sqlbuilder"
)

func TestConditions(t *testing.T) {
	t.Run("Chain Condition", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("a").Eq(1).
				And(nil).
				And(Col("b").LeftLike("text")).
				Or(Col("a").Eq(2)).
				Xor(Col("b").RightLike("g")),

			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
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
						Col("c").In(1, 2),
						Col("c").In([]int{3, 4}),
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),

			"(((c IN (?,?)) AND (c IN (?,?)) AND (a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, 2, 3, 4, 1, "%text%", 2, "%g%",
		)
	})
	t.Run("skip nil", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		)
	})
	t.Run("XOR and OR", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				Col("a").In(),
				Or(
					Col("a").NotIn(),
					And(
						nil,
						Col("a").Eq(1),
						Col("b").Like("text"),
					),
					Col("a").Eq(2),
				),
				Col("b").Like("g"),
			),
			"(((a = ?) AND (b LIKE ?)) OR (a = ?)) XOR (b LIKE ?)",
			1, "%text%", 2, "%g%",
		)
	})
	t.Run("XOR", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Xor(
				Col("a").Eq(1),
				Col("b").Like("g"),
			),
			"(a = ?) XOR (b LIKE ?)",
			1, "%g%",
		)
	})
	t.Run("Like", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Like("e"),
			"d LIKE ?",
			"%e%",
		)
	})
	t.Run("Not like", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").NotLike("e"),
			"d NOT LIKE ?",
			"%e%",
		)
	})
	t.Run("Equal", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Eq("e"),
			"d = ?", "e",
		)
	})
	t.Run("Not Equal", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Neq("e"),
			"d <> ?",
			"e",
		)
	})
	t.Run("In", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").In("e", "f"),
			"d IN (?,?)", "e", "f",
		)
	})
	t.Run("NotIn", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").NotIn("e", "f"),
			"d NOT IN (?,?)", "e", "f",
		)
	})
	t.Run("Less than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Lt(3),
			"d < ?", 3,
		)
	})
	t.Run("Less or equal than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Lte(3),
			"d <= ?", 3,
		)
	})
	t.Run("Greater than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Gt(3),
			"d > ?", 3,
		)
	})
	t.Run("Greater or equal than", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Gte(3),
			"d >= ?", 3,
		)
	})
	t.Run("Between", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").Between(0, 2),
			"d BETWEEN ? AND ?", 0, 2,
		)
	})
	t.Run("Not between", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").NotBetween(0, 2),
			"d NOT BETWEEN ? AND ?", 0, 2,
		)
	})
	t.Run("Is null", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").IsNull(),
			"d IS NULL",
		)
	})
	t.Run("Is not null", func(t *testing.T) {
		testutil.ShouldBeExpr(t,
			Col("d").IsNotNull(),
			"d IS NOT NULL",
		)
	})
}
