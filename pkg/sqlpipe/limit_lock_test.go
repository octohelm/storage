package sqlpipe

import (
	"context"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/testdata/model"
)

func TestLimitAndLockBranches(t *testing.T) {
	src := FromAll[model.User]()

	Then(
		t, "负数 limit 直接返回原 source",
		Expect(Limit[model.User](-1).Next(src) == src, Equal(true)),
	)

	limited := src.Pipe(Limit[model.User](10), Limit[model.User](5, Offset(2)))
	q, _ := sqlfrag.Collect(context.Background(), limited)
	Then(
		t, "重复 limit 会覆盖数值并附加 offset",
		Expect(strings.Contains(q, "LIMIT 5 OFFSET 2"), Equal(true)),
	)

	forUpdate, _ := sqlfrag.Collect(context.Background(), src.Pipe(ForUpdate[model.User](SkipLocked())))
	forNoKey, _ := sqlfrag.Collect(context.Background(), src.Pipe(ForNoKeyUpdate[model.User](NoWait())))
	forShare, _ := sqlfrag.Collect(context.Background(), src.Pipe(ForShare[model.User]()))
	forKeyShare, _ := sqlfrag.Collect(context.Background(), src.Pipe(ForKeyShare[model.User]()))

	Then(
		t, "不同锁模式生成各自 SQL 片段",
		Expect(strings.Contains(forUpdate, "FOR UPDATE SKIP LOCKED"), Equal(true)),
		Expect(strings.Contains(forNoKey, "FOR NO KEY UPDATE NO WAIT"), Equal(true)),
		Expect(strings.Contains(forShare, "FOR SHARE"), Equal(true)),
		Expect(strings.Contains(forKeyShare, "FOR KEY SHARE"), Equal(true)),
	)
}
