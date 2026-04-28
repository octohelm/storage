package internal

import (
	"context"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	sqlbuildermodelscoped "github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal/flags"
	"github.com/octohelm/storage/testdata/model"
)

func TestMutationAndBuilder(t *testing.T) {
	user := &model.User{Name: "alice"}

	insert := (&Builder[model.User]{}).BuildStmt(context.Background())
	q, args := sqlfrag.Collect(context.Background(), insert)
	Then(t, "默认 Builder 生成 select 语句并自动附加软删过滤",
		Expect(q, Equal("SELECT *\nFROM t_user\nWHERE f_deleted_at = ?")),
		Expect(args, Equal([]any{int64(0)})),
	)

	insert = (&Builder[model.User]{
		Source: &Mutation[model.User]{
			Values: func(yield func(*model.User) bool) {
				yield(user)
			},
			Strict: Strict[model.User]{Columns: []sqlbuildermodelscoped.Column[model.User]{model.UserT.Name}},
		},
	}).BuildStmt(context.Background())
	q, args = sqlfrag.Collect(context.Background(), insert)
	Then(t, "Insert 根据 Strict 列构造 VALUES",
		Expect(q, Equal("INSERT INTO t_user (f_name)\nVALUES\n\t(?)")),
		Expect(args, Equal([]any{"alice"})),
	)

	softDelete := (&Builder[model.User]{
		Source: &Mutation[model.User]{ForDelete: DeleteTypeSoft},
	}).BuildStmt(context.Background())
	q, args = sqlfrag.Collect(context.Background(), softDelete)
	Then(t, "软删除构造 update deleted_at 语句",
		Expect(q, Equal("UPDATE t_user\nSET f_deleted_at = ?\nWHERE f_deleted_at = ?")),
		Expect(len(args), Equal(2)),
		Expect(args[1] == int64(0), Equal(true)),
	)

	hardDelete := (&Builder[model.User]{
		Source: &Mutation[model.User]{ForDelete: DeleteTypeHard},
	}).BuildStmt(context.Background())
	q, _ = sqlfrag.Collect(context.Background(), hardDelete)
	Then(t, "硬删除保持 delete 语句",
		Expect(q, Equal("DELETE FROM t_user\nWHERE f_deleted_at = ?")),
	)

	update := (&Builder[model.User]{
		Source: &Mutation[model.User]{
			ForUpdate: true,
			Values: func(yield func(*model.User) bool) {
				yield(&model.User{Name: "bob"})
			},
			Strict: Strict[model.User]{Columns: []sqlbuildermodelscoped.Column[model.User]{model.UserT.Name}},
		},
	}).BuildStmt(context.Background())
	q, args = sqlfrag.Collect(context.Background(), update)
	Then(t, "Update 根据赋值列表构造 set 子句",
		Expect(q, Equal("UPDATE t_user\nSET f_name = ?")),
		Expect(args, Equal([]any{"bob"})),
	)
}

func TestBuilderBranches(t *testing.T) {
	emptySelect := (&Builder[model.User]{
		Flag: flags.WhereOrPagerRequired | flags.IncludesAll,
	}).BuildStmt(context.Background())
	q, _ := sqlfrag.Collect(context.Background(), emptySelect)
	Then(t, "WhereOrPagerRequired 在无 where 和 pager 时返回空语句",
		Expect(q, Equal("")),
	)

	selected := (&Builder[model.User]{}).
		WithDefaultProjects(model.UserT.Name).
		WithDistinctOn(model.UserT.Name).
		WithOrders(sqlbuilder.DescOrder(model.UserT.ID)).
		WithPager(sqlbuilder.Limit(10)).
		BuildStmt(context.Background())
	q, args := sqlfrag.Collect(context.Background(), selected)
	Then(t, "DistinctOn 会补齐默认排序并附加分页",
		Expect(strings.Contains(q, "SELECT DISTINCT ON ( f_name ) f_name"), Equal(true)),
		Expect(strings.Contains(q, "FROM t_user"), Equal(true)),
		Expect(strings.Contains(q, "WHERE f_deleted_at = ?"), Equal(true)),
		Expect(strings.Contains(q, "LIMIT 10"), Equal(true)),
		Expect(args, Equal([]any{int64(0)})),
	)

	noPager := (&Builder[model.User]{Flag: flags.WithoutPager | flags.WithoutSorter}).
		WithProjects(model.UserT.Name).
		WithOrders(sqlbuilder.DescOrder(model.UserT.ID)).
		WithPager(sqlbuilder.Limit(10)).
		BuildStmt(context.Background())
	q, args = sqlfrag.Collect(context.Background(), noPager)
	Then(t, "WithoutPager 和 WithoutSorter 会抑制分页与排序附加项",
		Expect(q, Equal("SELECT f_name\nFROM t_user\nWHERE f_deleted_at = ?")),
		Expect(args, Equal([]any{int64(0)})),
	)
}
