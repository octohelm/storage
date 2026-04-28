package sqlfrag

import (
	"context"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestConstAndEmptyInternal(t *testing.T) {
	q, args := Collect(context.Background(), Const(""))
	emptyQ, emptyArgs := Collect(context.Background(), Empty())

	Then(t, "空 Const 与 Empty 都可安全收集",
		Expect(Const("").IsNil(), Equal(true)),
		Expect(Const("x").IsNil(), Equal(false)),
		Expect(q, Equal("")),
		Expect(len(args), Equal(0)),
		Expect(emptyQ, Equal("")),
		Expect(len(emptyArgs), Equal(0)),
	)
}
