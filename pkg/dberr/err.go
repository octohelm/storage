package dberr

import "fmt"

func New(tpe errType, msg string) *SqlError {
	return &SqlError{
		Type: tpe,
		Msg:  msg,
	}
}

type SqlError struct {
	Type errType
	Msg  string
}

func (e *SqlError) Error() string {
	return fmt.Sprintf("Sqlx [%s] %s", e.Type, e.Msg)
}

type errType string

var (
	ErrTypeNotFound   errType = "NotFound"
	ErrTypeConflict   errType = "Conflict"
	ErrTypeRolledBack errType = "RolledBack"
)

func IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := UnwrapAll(err).(*SqlError); ok {
		return sqlErr.Type == ErrTypeNotFound
	}
	return false
}

func IsErrConflict(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := UnwrapAll(err).(*SqlError); ok {
		return sqlErr.Type == ErrTypeConflict
	}
	return false
}

func IsErrRolledBack(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := UnwrapAll(err).(*SqlError); ok {
		return sqlErr.Type == ErrTypeRolledBack
	}
	return false
}
