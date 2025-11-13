package parser

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"text/scanner"

	duckdbtoken "github.com/octohelm/storage/internal/sql/adapter/duckdb/token"
)

func Parse(sql []byte) (any, error) {
	p := &parser{}
	p.Init(bytes.NewBuffer(sql))

	for expr := range p.Tokens() {
		fmt.Println(expr)
	}

	return nil, nil
}

type parser struct {
	scanner.Scanner
}

func (p *parser) Init(r io.Reader) {
	p.Scanner.Init(r)
	p.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanFloats | scanner.ScanInts | scanner.ScanComments
	p.Error = func(s *scanner.Scanner, msg string) {}
}

type Token string

func (t Token) Kind() duckdbtoken.Token {
	if len(t) > 0 {
		switch t[0] {
		case '\'', '"':
			return duckdbtoken.STRING
		case '0':
			return duckdbtoken.INT
		case 'f':
			return duckdbtoken.FLOAT
		default:
			return duckdbtoken.IDENT
		}
	}
	return duckdbtoken.ILLEGAL
}

func (t Token) Value() string {
	return string(t)
}

func (p *parser) Tokens() iter.Seq[Token] {
	return func(yield func(Token) bool) {

	}
}
