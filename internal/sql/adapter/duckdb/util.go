package duckdb

import (
	"io"
	textscanner "text/scanner"
)

type stmtType int

// https://duckdb.org/docs/stable/sql/statements/create_table#syntax
const (
	stmtTypeInvalid stmtType = iota
	stmtTypeName
	stmtTypePrimaryKey
	stmtTypeNotNull
	stmtTypeDefault
)

func parseTableDecls(r io.Reader) map[string][]string {
	s := &textscanner.Scanner{}
	s.Init(r)
	s.Error = func(s *textscanner.Scanner, msg string) {}

	scope := 0
	cols := make(map[string][]string)

	decl := make([]string, 0)

	commitDecl := func() {
		if len(decl) == 0 || scope != 1 {
			return
		}
		cols[decl[0]] = decl[1:]
		decl = make([]string, 0)
	}

	appendOrConcat := func(part string) {
		if len(decl) > 0 {

			if cols[part] == nil {
			}

		}

		decl = append(decl, part)

	}

	for tok := s.Scan(); tok != textscanner.EOF; tok = s.Scan() {
		part := s.TokenText()

		switch part {
		case "(":
			scope++
			if scope == 1 {
				continue
			}
		case ")":
			if scope == 1 {
				commitDecl()
				continue
			}
			scope--
		case ",":
			if scope == 1 {
				commitDecl()
				continue
			}
		}

		if scope > 0 {
			appendOrConcat(part)
		}
	}

	return cols
}
