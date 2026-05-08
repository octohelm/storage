package directive

import (
	"net/http"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestDirectiveErrors(t *testing.T) {
	Then(
		t, "无名非法指令返回通用消息和 400",
		Expect((&ErrInvalidDirective{}).StatusCode(), Equal(http.StatusBadRequest)),
		Expect((&ErrInvalidDirective{}).Error(), Equal("invalid directive")),
	)

	Then(
		t, "命名非法指令和不支持指令返回明确错误",
		Expect((&ErrInvalidDirective{DirectiveName: "where"}).Error(), Equal("invalid directive: where")),
		Expect((&ErrUnsupportedDirective{DirectiveName: "eq"}).StatusCode(), Equal(http.StatusBadRequest)),
		Expect((&ErrUnsupportedDirective{DirectiveName: "eq"}).Error(), Equal("unsupported directive: eq")),
	)
}
