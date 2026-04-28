package sqlpipe

import (
	"context"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestJoinVariants(t *testing.T) {
	full, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		FullJoinOnAs[joinedOrgUser](model.OrgUserT.UserID, model.UserT.ID),
	))
	cross, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		CrossJoinOnAs[joinedOrgUser](model.OrgUserT.UserID, model.UserT.ID),
	))
	inner, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		InnerJoinOnAs[joinedOrgUser](model.OrgUserT.UserID, model.UserT.ID),
	))
	left, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		LeftJoinOnAs[joinedOrgUser](model.OrgUserT.OrgID, model.OrgT.ID),
	))
	right, _ := sqlfrag.Collect(context.Background(), FromAll[joinedOrgUser]().Pipe(
		RightJoinOnAs[joinedOrgUser](model.OrgUserT.OrgID, model.OrgT.ID),
	))

	Then(t, "不同 join 构造器生成对应关键字",
		Expect(strings.Contains(full, "FULL JOIN t_user"), Equal(true)),
		Expect(strings.Contains(cross, "CROSS JOIN t_user"), Equal(true)),
		Expect(strings.Contains(inner, "INNER JOIN t_user"), Equal(true)),
		Expect(strings.Contains(left, "LEFT JOIN t_org"), Equal(true)),
		Expect(strings.Contains(right, "RIGHT JOIN t_org"), Equal(true)),
	)
}
