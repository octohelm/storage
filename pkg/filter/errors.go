package filter

import (
	"fmt"
	"net/http"
)

type errBadRequest struct{}

func (errBadRequest) StatusCode() int {
	return http.StatusBadRequest
}

// ErrInvalidFilterOp 表示过滤操作符非法。
type ErrInvalidFilterOp struct {
	errBadRequest

	Op string
}

func (e *ErrInvalidFilterOp) Error() string {
	return fmt.Sprintf("invalid filter op `%s`", e.Op)
}

// ErrInvalidFilter 表示过滤表达式非法。
type ErrInvalidFilter struct {
	errBadRequest

	Filter string
}

func (e *ErrInvalidFilter) Error() string {
	return fmt.Sprintf("invalid filter `%s`", e.Filter)
}

// ErrUnsupportedQLField 表示查询字段不受支持。
type ErrUnsupportedQLField struct {
	errBadRequest

	FieldName string
}

func (e *ErrUnsupportedQLField) Error() string {
	return fmt.Sprintf("unsupported ql field `%s`", e.FieldName)
}
