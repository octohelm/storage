package sqlpipe

import (
	"context"
	"slices"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

type joinedOrgUser struct {
	model.OrgUser
	User model.User
	Org  model.Org
}

func TestProjectSortAndHelpers(t *testing.T) {
	base := FromAll[model.User]()

	Then(t, "Columns helper 透传列序列",
		Expect(len(slices.Collect(Columns(model.UserT.Name, model.UserT.Age))), Equal(2)),
	)

	selected, _ := sqlfrag.Collect(context.Background(), base.Pipe(
		Project[model.User](sqlfrag.Const("1")),
	))
	defaultSelected, _ := sqlfrag.Collect(context.Background(), base.Pipe(
		DefaultProject[model.User](model.UserT.Name),
	))
	castSelected, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		CastSelect[joinedOrgUser](model.UserT.Name),
	))
	sorted, _ := sqlfrag.Collect(context.Background(), base.Pipe(
		CastAscSort[model.User](model.OrgT.Name),
		CastDescSort[model.User](model.OrgT.ID),
	))

	Then(t, "Project、DefaultProject、CastSelect 和 Sort 会生成预期片段",
		Expect(strings.Contains(selected, "SELECT 1"), Equal(true)),
		Expect(strings.Contains(defaultSelected, "SELECT f_name"), Equal(true)),
		Expect(strings.Contains(castSelected, "SELECT f_name"), Equal(true)),
		Expect(strings.Contains(sorted, "ORDER BY (f_name) ASC,(f_id) DESC"), Equal(true)),
	)

	ascQ, _ := sqlfrag.Collect(context.Background(), modelscoped.AscOrder(model.UserT.Name))
	descQ, _ := sqlfrag.Collect(context.Background(), modelscoped.DescOrder(model.UserT.ID))
	Then(t, "modelscoped 排序 helper 透传到底层 sqlbuilder.Order",
		Expect(ascQ, Equal("(f_name) ASC")),
		Expect(descQ, Equal("(f_id) DESC")),
	)
}
