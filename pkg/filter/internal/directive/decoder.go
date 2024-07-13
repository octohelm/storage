package directive

import (
	"bytes"
	"io"
	"text/scanner"
)

type Marshaler interface {
	MarshalDirective() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalDirective(*Decoder) error
}

type UnmarshalerFunc func(*Decoder) error

func (fn UnmarshalerFunc) UnmarshalDirective(d *Decoder) error {
	return fn(d)
}

func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{}
	d.unmarshalers = map[string]Newer{}
	d.Reset(r)
	return d
}

type Decoder struct {
	unmarshalers map[string]Newer
	s            scanner.Scanner

	directiveName string

	kind Kind
	text []byte

	tmp bytes.Buffer
}

type Newer func() Unmarshaler

const DefaultDirectiveNewer = "_default"

func (d *Decoder) RegisterDirectiveNewer(directiveName string, fn Newer) {
	d.unmarshalers[directiveName] = fn
}

func (d *Decoder) Unmarshaler(name string) (Unmarshaler, error) {
	d.directiveName = name

	if u, ok := d.unmarshalers[name]; ok {
		return u(), nil
	}

	if u, ok := d.unmarshalers[DefaultDirectiveNewer]; ok {
		return u(), nil
	}

	return nil, &ErrUnsupportedDirective{
		DirectiveName: name,
	}
}

func (d *Decoder) Reset(r io.Reader) {
	d.s.Init(r)
	d.tmp.Reset()
}

func (d *Decoder) DirectiveName() (string, error) {
	if d.directiveName == "" {
		k, _ := d.Next()
		if k != KindFuncStart {
			return "", &ErrInvalidDirective{}
		}
	}
	return d.directiveName, nil
}

func (d *Decoder) textToken() []byte {
	textToken := d.tmp.Bytes()
	d.tmp.Reset()
	return textToken
}

func (d *Decoder) Next() (Kind, []byte) {
	tok := d.s.Scan()
	switch tok {
	case scanner.EOF:
		return EOF, nil
	case ')':
		return KindFuncEnd, d.textToken()
	case ',':
		return d.Next()
	default:
		tokenText := d.s.TokenText()

		switch d.s.Peek() {
		case '(':
			d.directiveName = tokenText

			return KindFuncStart, []byte(tokenText)
		case ',', ')':
			return KindValue, []byte(tokenText)
		}
	}
	return d.Next()
}

type Kind int

const (
	KindInvalid Kind = iota
	KindFuncStart
	KindFuncEnd
	KindValue
	EOF
)

type RawValue []byte

func (v RawValue) MarshalDirective() ([]byte, error) {
	return v, nil
}
