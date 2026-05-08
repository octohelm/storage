package modelscoped_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlbuilder/modelscoped"
	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestOrderHelpers(t *testing.T) {
	ascQ, _ := sqlfrag.Collect(context.Background(), modelscoped.AscOrder(model.UserT.Name))
	descQ, _ := sqlfrag.Collect(context.Background(), modelscoped.DescOrder(model.UserT.ID))

	Then(
		t, "modelscoped 排序 helper 生成底层 order SQL",
		Expect(ascQ, Equal("(f_name) ASC")),
		Expect(descQ, Equal("(f_id) DESC")),
	)
}
