package dberr

import "fmt"

func NewSqlError(tpe sqlErrType, msg string) *SqlError {
	return &SqlError{
		Type: tpe,
		Msg:  msg,
	}
}

type SqlError struct {
	Type sqlErrType
	Msg  string
}

func (e *SqlError) Error() string {
	return fmt.Sprintf("Sqlx [%s] %s", e.Type, e.Msg)
}

type sqlErrType string

var (
	SqlErrTypeNotFound sqlErrType = "NotFound"
	SqlErrTypeConflict sqlErrType = "Conflict"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := UnwrapAll(err).(*SqlError); ok {
		return sqlErr.Type == SqlErrTypeNotFound
	}
	return true
}

func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := UnwrapAll(err).(*SqlError); ok {
		return sqlErr.Type == SqlErrTypeConflict
	}
	return true
}
