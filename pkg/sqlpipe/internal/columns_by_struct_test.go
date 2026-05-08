package internal

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestColumnsByStruct(t *testing.T) {
	type joined struct {
		model.User
		Org model.Org `alias:"t_org"`
	}

	q, _ := sqlfrag.Collect(context.Background(), ColumnsByStruct(joined{}))
	Then(
		t, "ColumnsByStruct 生成结构体字段投影",
		Expect(q, Equal("t_user.f_id AS t_user__f_id, t_user.f_name AS t_user__f_name, t_user.f_nickname AS t_user__f_nickname, t_user.f_username AS t_user__f_username, t_user.f_gender AS t_user__f_gender, t_user.f_age AS t_user__f_age, t_user.f_created_at AS t_user__f_created_at, t_user.f_updated_at AS t_user__f_updated_at, t_user.f_deleted_at AS t_user__f_deleted_at, t_org.f_id AS t_org__f_id, t_org.f_name AS t_org__f_name, t_org.f_created_at AS t_org__f_created_at, t_org.f_updated_at AS t_org__f_updated_at, t_org.f_deleted_at AS t_org__f_deleted_at")),
	)
}
