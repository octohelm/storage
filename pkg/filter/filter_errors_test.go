package filter_test

import (
	"net/http"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	filter "github.com/octohelm/storage/pkg/filter"
)

func TestErrors(t *testing.T) {
	Then(
		t, "错误类型暴露 400 状态码和明确消息",
		Expect((&filter.ErrInvalidFilterOp{Op: "x"}).StatusCode(), Equal(http.StatusBadRequest)),
		Expect((&filter.ErrInvalidFilterOp{Op: "x"}).Error(), Equal("invalid filter op `x`")),
		Expect((&filter.ErrInvalidFilter{Filter: "f"}).Error(), Equal("invalid filter `f`")),
		Expect((&filter.ErrUnsupportedQLField{FieldName: "name"}).Error(), Equal("unsupported ql field `name`")),
	)
}
