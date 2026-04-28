package nullable

import (
	"database/sql"
)

import (
	_ "unsafe"
)

// NewNullIgnoreScanner 包装目标值，并在 Scan 时忽略 nil 源值。
func NewNullIgnoreScanner(dest any) *NullIgnoreScanner {
	return &NullIgnoreScanner{
		dest: dest,
	}
}

// NullIgnoreScanner 用于在目标不希望被零值覆盖时跳过 nil 数据库值。
type NullIgnoreScanner struct {
	dest any
}

func (scanner *NullIgnoreScanner) Scan(src any) error {
	if s, ok := scanner.dest.(sql.Scanner); ok {
		return s.Scan(src)
	}
	if src == nil {
		return nil
	}
	return convertAssign(scanner.dest, src)
}

//go:linkname convertAssign database/sql.convertAssign
func convertAssign(dest, src any) error
