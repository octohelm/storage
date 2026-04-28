package directive

import (
	"fmt"
	"net/http"
)

// ErrInvalidDirective 表示 directive 文本不合法。
type ErrInvalidDirective struct {
	DirectiveName string
}

func (e *ErrInvalidDirective) StatusCode() int {
	return http.StatusBadRequest
}

func (e *ErrInvalidDirective) Error() string {
	if e.DirectiveName == "" {
		return fmt.Sprintf("invalid directive")
	}

	return fmt.Sprintf("invalid directive: %s", e.DirectiveName)
}

// ErrUnsupportedDirective 表示 directive 名称不受支持。
type ErrUnsupportedDirective struct {
	DirectiveName string
}

func (e *ErrUnsupportedDirective) StatusCode() int {
	return http.StatusBadRequest
}

func (e *ErrUnsupportedDirective) Error() string {
	return fmt.Sprintf("unsupported directive: %s", e.DirectiveName)
}
