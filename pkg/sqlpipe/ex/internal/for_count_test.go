package internal_test

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe"
	internal "github.com/octohelm/storage/pkg/sqlpipe/ex/internal"
	"github.com/octohelm/storage/testdata/model"
)

func TestForCount(t *testing.T) {
	src := sqlpipe.FromAll[model.User]().Pipe(
		sqlpipe.Limit[model.User](10),
		internal.ForCount[model.User](),
	)

	q, _ := sqlfrag.Collect(context.Background(), src)
	Then(t, "ForCount 去掉排序和分页标记",
		Expect(q, Equal("SELECT *\nFROM t_user")),
	)
}
