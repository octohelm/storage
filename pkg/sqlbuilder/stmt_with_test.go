package sqlbuilder_test

import (
	"testing"

	"github.com/octohelm/storage/pkg/sqlfrag"

	"github.com/octohelm/storage/pkg/sqlfrag/testutil"

	. "github.com/octohelm/storage/pkg/sqlbuilder"
	testingx "github.com/octohelm/x/testing"
)

func TestWithStmt(t *testing.T) {
	gr := &GroupRelation{}
	g := &Group{}

	t.Run("simple with", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,

			With((&GroupWithParent{}).T(), func(tmpTableGroupWithParent Table) sqlfrag.Fragment {
				s := Select(MultiMayAutoAlias(
					TypedColOf[int](g.T(), "f_group_id"),
					TypedColOf[int](gr.T(), "f_group_id"),
				)).
					From(gr.T(),
						RightJoin(g.T()).On(
							TypedColOf[int](g.T(), "f_group_id").V(EqCol(TypedColOf[int](gr.T(), "f_group_id")))),
					)
				return s
			}).
				With((&GroupWithParent{}).T(), func(tmpTableGroupWithParent Table) sqlfrag.Fragment {
					s := Select(MultiMayAutoAlias(
						TypedColOf[int](g.T(), "f_group_id"),
						TypedColOf[int](gr.T(), "f_group_id"),
					)).
						From(gr.T(),
							RightJoin(g.T()).On(
								TypedColOf[int](g.T(), "f_group_id").V(
									EqCol(TypedColOf[int](gr.T(), "f_group_id")),
								),
							),
						)
					return s
				}).
				Exec(func(tables ...Table) sqlfrag.Fragment {
					return Select(nil).From(tables[0])
				}),
			testutil.BeFragment(`
WITH
t_group_with_parent(f_group_id,f_parent_group_id) AS (
	SELECT t_group.f_group_id AS t_group__f_group_id, t_group_relation.f_group_id AS t_group_relation__f_group_id
	FROM t_group_relation
	RIGHT JOIN t_group ON t_group.f_group_id = t_group_relation.f_group_id
),
t_group_with_parent(f_group_id,f_parent_group_id) AS (
	SELECT t_group.f_group_id AS t_group__f_group_id, t_group_relation.f_group_id AS t_group_relation__f_group_id
	FROM t_group_relation
	RIGHT JOIN t_group ON t_group.f_group_id = t_group_relation.f_group_id
)
SELECT *
FROM t_group_with_parent
`))
	})

	t.Run("WithRecursive", func(t *testing.T) {
		testingx.Expect[sqlfrag.Fragment](t,
			WithRecursive((&GroupWithParentAndChildren{}).T(), func(tmpTableGroupWithParentAndChildren Table) sqlfrag.Fragment {
				return With((&GroupWithParent{}).T(), func(tmpTableGroupWithParent Table) sqlfrag.Fragment {
					s := Select(MultiMayAutoAlias(
						TypedColOf[int](g.T(), "f_group_id"),
						TypedColOf[int](gr.T(), "f_parent_group_id"),
					)).
						From(gr.T(), RightJoin(g.T()).On(
							TypedColOf[int](g.T(), "f_group_id").V(
								EqCol(TypedColOf[int](gr.T(), "f_group_id"))),
						))
					return s
				}).Exec(func(tables ...Table) sqlfrag.Fragment {
					tmpTableGroupWithParent := tables[0]

					return Select(
						MultiMayAutoAlias(
							tmpTableGroupWithParent.F("f_group_id"),
							tmpTableGroupWithParent.F("f_parent_group_id"),
							Alias(sqlfrag.Pair("0"), "f_depth"),
						),
					).From(
						tmpTableGroupWithParent,
						Where(
							TypedColOf[int](tmpTableGroupWithParent, "f_group_id").V(Eq(1201375536060956676)),
						),
						Union().All(
							Select(MultiMayAutoAlias(
								tmpTableGroupWithParent.F("f_group_id"),
								tmpTableGroupWithParent.F("f_parent_group_id"),
								Alias(tmpTableGroupWithParentAndChildren.F("f_depth").Fragment("# + 1"), "f_depth"),
							)).From(
								tmpTableGroupWithParent,
								CrossJoin(tmpTableGroupWithParentAndChildren),
								Where(
									And(
										TypedColOf[int](tmpTableGroupWithParent, "f_group_id").V(NeqCol(TypedColOf[int](tmpTableGroupWithParentAndChildren, "f_group_id"))),
										TypedColOf[int](tmpTableGroupWithParent, "f_parent_group_id").V(EqCol(TypedColOf[int](tmpTableGroupWithParentAndChildren, "f_group_id"))),
									),
								),
							),
						),
					)
				})
			}).Exec(func(tables ...Table) sqlfrag.Fragment {
				return Select(nil).From(tables[0])
			}),
			testutil.BeFragment(`
WITH RECURSIVE
t_group_with_parent_and_children(f_group_id,f_parent_group_id,f_depth) AS (
	WITH
	t_group_with_parent(f_group_id,f_parent_group_id) AS (
		SELECT t_group.f_group_id AS t_group__f_group_id, t_group_relation.f_parent_group_id AS t_group_relation__f_parent_group_id
		FROM t_group_relation
		RIGHT JOIN t_group ON t_group.f_group_id = t_group_relation.f_group_id
	)
	SELECT f_group_id, f_parent_group_id, 0 AS f_depth
	FROM t_group_with_parent
	WHERE f_group_id = ?
	UNION ALL 
	SELECT t_group_with_parent.f_group_id AS t_group_with_parent__f_group_id, t_group_with_parent.f_parent_group_id AS t_group_with_parent__f_parent_group_id, t_group_with_parent_and_children.f_depth + 1 AS f_depth
	FROM t_group_with_parent
	CROSS JOIN t_group_with_parent_and_children
	WHERE (t_group_with_parent.f_group_id <> t_group_with_parent_and_children.f_group_id) AND (t_group_with_parent.f_parent_group_id = t_group_with_parent_and_children.f_group_id)
)
SELECT *
FROM t_group_with_parent_and_children
`, 1201375536060956676))
	})
}

var tableGroup = TableFromModel(&Group{})

type Group struct {
	GroupID int `db:"f_group_id"`
}

func (g *Group) TableName() string {
	return "t_group"
}

func (g *Group) T() Table {
	return tableGroup
}

var tableGroupRelation = TableFromModel(&GroupRelation{})

type GroupRelation struct {
	GroupID       int `db:"f_group_id"`
	ParentGroupID int `db:"f_parent_group_id"`
}

func (g *GroupRelation) TableName() string {
	return "t_group_relation"
}

func (g *GroupRelation) T() Table {
	return tableGroupRelation
}

var tableGroupWithParent = TableFromModel(&GroupWithParent{})

type GroupWithParent struct {
	GroupID       int `db:"f_group_id"`
	ParentGroupID int `db:"f_parent_group_id"`
}

func (g *GroupWithParent) TableName() string {
	return "t_group_with_parent"
}

func (g *GroupWithParent) T() Table {
	return tableGroupWithParent
}

var tableGroupWithParentAndChildren = TableFromModel(&GroupWithParentAndChildren{})

type GroupWithParentAndChildren struct {
	GroupWithParent
	Depth int `db:"f_depth"`
}

func (g *GroupWithParentAndChildren) TableName() string {
	return "t_group_with_parent_and_children"
}

func (g *GroupWithParentAndChildren) T() Table {
	return tableGroupWithParentAndChildren
}
