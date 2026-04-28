package sqlpipe

import (
	"context"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
	"github.com/octohelm/storage/pkg/sqlpipe/internal"
	"github.com/octohelm/storage/testdata/model"
)

func TestAliasAndNoopMethods(t *testing.T) {
	base := FromAll[model.User]()
	aliased, _ := sqlfrag.Collect(context.Background(), As(base.Pipe(Limit[model.User](1)), "u"))

	Then(t, "Alias source 会生成别名片段",
		Expect(strings.Contains(aliased, ") AS u"), Equal(true)),
	)

	n := &noop[model.User]{}
	q, _ := sqlfrag.Collect(context.Background(), n)

	Then(t, "noop 的 Frag、Pipe、ApplyStmt 都保持空操作",
		Expect(q, Equal("")),
		Expect(n.Pipe() == n, Equal(true)),
		Expect(n.ApplyStmt(context.Background(), &internal.Builder[model.User]{}) != nil, Equal(true)),
	)

	alias := As(base.Pipe(Limit[model.User](1)), "u")
	Then(t, "Alias source 可继续输出 SQL 字符串",
		Expect(alias.IsNil(), Equal(false)),
		Expect(alias.(interface{ String() string }).String() != "", Equal(true)),
	)

	values := Value(&model.User{Name: "alice"}, model.UserT.Name).(*sourceValues[model.User])
	vq, _ := sqlfrag.Collect(context.Background(), values)
	Then(t, "sourceValues 暴露 String、Frag 和 ApplyStmt 行为",
		Expect(values.IsNil(), Equal(false)),
		Expect(values.String() != "", Equal(true)),
		Expect(vq != "", Equal(true)),
		Expect(values.ApplyStmt(context.Background(), &internal.Builder[model.User]{}).Source != nil, Equal(true)),
	)
}
