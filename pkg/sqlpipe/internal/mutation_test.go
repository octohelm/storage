package internal

import (
	"context"
	"slices"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder"
	sqlbuildermodelscoped "github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestMutationStrict(t *testing.T) {
	tbl := sqlbuilder.TableFromModel(&model.User{})
	mut := &Mutation[model.User]{
		Strict: Strict[model.User]{
			Omit:    true,
			Columns: []sqlbuildermodelscoped.Column[model.User]{model.UserT.Name},
		},
	}

	cols := mut.Strict.StrictColumnCollection(tbl)
	names := make([]string, 0)
	for col := range cols.Cols() {
		names = append(names, col.FieldName())
	}

	Then(t, "Strict.Omit 生成排除列集合",
		Expect(slices.Contains(names, "Name"), Equal(false)),
		Expect(slices.Contains(names, "Nickname"), Equal(true)),
	)
}

func TestMutationBranches(t *testing.T) {
	tbl := sqlbuilder.TableFromModel(&model.User{})

	assignments := slices.Collect((&Mutation[model.User]{
		Values: func(yield func(*model.User) bool) {
			yield(&model.User{Name: "alice"})
		},
		OmitZero: OmitZero[model.User]{
			Enabled: true,
			Exclude: []sqlbuildermodelscoped.Column[model.User]{model.UserT.Name},
		},
	}).PrepareAssignments(context.Background(), tbl))

	Then(t, "OmitZero 赋值会保留显式排除列和非零字段",
		Expect(len(assignments) >= 1, Equal(true)),
	)

	insertOmitZero := (&Builder[model.User]{
		Source: &Mutation[model.User]{
			Values: func(yield func(*model.User) bool) {
				yield(&model.User{Name: "alice"})
			},
			OmitZero: OmitZero[model.User]{
				Enabled: true,
				Exclude: []sqlbuildermodelscoped.Column[model.User]{model.UserT.Name},
			},
		},
	}).BuildStmt(context.Background())
	q, args := sqlfrag.Collect(context.Background(), insertOmitZero)
	Then(t, "Insert OmitZero 会根据首个值推导列集合",
		Expect(strings.Contains(q, "INSERT INTO t_user (f_name,f_created_at)"), Equal(true)),
		Expect(len(args), Equal(2)),
		Expect(args[0], Equal(any("alice"))),
	)

	insertFrom := (&Builder[model.User]{
		Source: &Mutation[model.User]{
			From: sqlbuilder.Select(model.UserT.Name).From(model.UserT),
		},
	}).BuildStmt(context.Background())
	q, _ = sqlfrag.Collect(context.Background(), insertFrom)
	Then(t, "Insert 支持从子查询写入",
		Expect(strings.Contains(q, "INSERT INTO t_user (f_name,f_nickname,f_username,f_gender,f_age,f_created_at,f_updated_at,f_deleted_at)"), Equal(true)),
		Expect(strings.Contains(q, "SELECT f_name"), Equal(true)),
		Expect(strings.Contains(q, "FROM t_user"), Equal(true)),
	)

	updateFrom := (&Builder[model.User]{
		Source: &Mutation[model.User]{
			ForUpdate: true,
			From:      &model.Org{},
			Assignments: []sqlbuilder.Assignment{
				sqlbuilder.ColumnsAndValues(model.UserT.Name, "alice"),
			},
		},
	}).BuildStmt(context.Background())
	q, args = sqlfrag.Collect(context.Background(), updateFrom)
	Then(t, "Update 支持 FROM 来源和显式赋值列表",
		Expect(q, Equal("UPDATE t_user\nSET f_name = ?\nFROM t_org")),
		Expect(args, Equal([]any{"alice"})),
	)
}
