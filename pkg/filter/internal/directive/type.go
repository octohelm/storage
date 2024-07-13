package directive

import (
	"fmt"
	"net/http"
)

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

type ErrUnsupportedDirective struct {
	DirectiveName string
}

func (e *ErrUnsupportedDirective) StatusCode() int {
	return http.StatusBadRequest
}

func (e *ErrUnsupportedDirective) Error() string {
	return fmt.Sprintf("unsupported directive: %s", e.DirectiveName)
}
