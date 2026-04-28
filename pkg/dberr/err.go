package dberr

import (
	"errors"
	"fmt"
)

// New 创建一个 SqlError。
func New(tpe errType, msg string) *SqlError {
	return &SqlError{
		Type: tpe,
		Msg:  msg,
	}
}

// SqlError 表示数据库层统一错误。
type SqlError struct {
	Type errType
	Msg  string
}

func (e *SqlError) Error() string {
	return fmt.Sprintf("SqlError{%s} %s", e.Type, e.Msg)
}

type errType string

var (
	ErrTypeNotFound   errType = "NotFound"
	ErrTypeConflict   errType = "Conflict"
	ErrTypeRolledBack errType = "RolledBack"
)

// IsErrNotFound 判断错误是否为未找到。
func IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := errors.AsType[*SqlError](err); ok {
		return sqlErr.Type == ErrTypeNotFound
	}
	return false
}

// IsErrConflict 判断错误是否为冲突。
func IsErrConflict(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := errors.AsType[*SqlError](err); ok {
		return sqlErr.Type == ErrTypeConflict
	}
	return false
}

// IsErrRolledBack 判断错误是否为事务回滚。
func IsErrRolledBack(err error) bool {
	if err == nil {
		return false
	}
	if sqlErr, ok := errors.AsType[*SqlError](err); ok {
		return sqlErr.Type == ErrTypeRolledBack
	}
	return false
}
