package filter

import (
	"fmt"
	"net/http"
)

type errBadRequest struct{}

func (errBadRequest) StatusCode() int {
	return http.StatusBadRequest
}

type ErrInvalidFilterOp struct {
	errBadRequest

	Op string
}

func (e *ErrInvalidFilterOp) Error() string {
	return fmt.Sprintf("invalid filter op `%s`", e.Op)
}

type ErrInvalidFilter struct {
	errBadRequest

	Filter string
}

func (e *ErrInvalidFilter) Error() string {
	return fmt.Sprintf("invalid filter `%s`", e.Filter)
}

type ErrUnsupportedQLField struct {
	errBadRequest

	FieldName string
}

func (e *ErrUnsupportedQLField) Error() string {
	return fmt.Sprintf("unsupported ql field `%s`", e.FieldName)
}
