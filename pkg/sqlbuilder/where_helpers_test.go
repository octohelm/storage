package sqlbuilder

import (
	"context"
	"testing"

	testingx "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/storage/pkg/sqlfrag"
)

func TestWhereInternal(t *testing.T) {
	w := Where(sqlfrag.Pair("a = ?", 1))
	q, args := sqlfrag.Collect(context.Background(), w)

	testingx.Then(t, "Where 会包装普通 fragment 并输出 WHERE 前缀",
		testingx.Expect(w.AdditionType(), testingx.Equal(AdditionWhere)),
		testingx.Expect(w.IsNil(), testingx.Equal(false)),
		testingx.Expect(q, testingx.Equal("WHERE a = ?")),
		testingx.Expect(args, testingx.Equal([]any{1})),
	)

	wrapped := Where(w)
	testingx.Then(t, "重复 Where 不会二次包装",
		testingx.Expect(wrapped, testingx.Equal(w)),
	)

	testingx.Then(t, "空 where 被视为 nil",
		testingx.Expect(Where(nil).IsNil(), testingx.Equal(true)),
	)
}
