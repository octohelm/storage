package nullable

import (
	"database/sql"
)

import (
	_ "unsafe"
)

func NewNullIgnoreScanner(dest any) *NullIgnoreScanner {
	return &NullIgnoreScanner{
		dest: dest,
	}
}

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
