package dberr

import (
	"fmt"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestSqlError(t *testing.T) {
	err := New(ErrTypeNotFound, "user")

	Then(t, "构造错误保留类型和消息",
		Expect(err.Error(), Equal("SqlError{NotFound} user")),
		Expect(IsErrNotFound(err), Equal(true)),
		Expect(IsErrConflict(err), Equal(false)),
		Expect(IsErrRolledBack(err), Equal(false)),
	)

	Then(t, "包装错误可通过错误链识别",
		Expect(IsErrConflict(fmt.Errorf("wrapped: %w", New(ErrTypeConflict, "id"))), Equal(true)),
		Expect(IsErrRolledBack(fmt.Errorf("wrapped: %w", New(ErrTypeRolledBack, "tx"))), Equal(true)),
	)

	Then(t, "空错误和非 SqlError 不命中",
		Expect(IsErrNotFound(nil), Equal(false)),
		Expect(IsErrConflict(fmt.Errorf("plain")), Equal(false)),
		Expect(IsErrRolledBack(fmt.Errorf("plain")), Equal(false)),
	)
}
