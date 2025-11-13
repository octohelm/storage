package token

type Token int

const (
	ILLEGAL Token = iota
	EOF
	COMMENT

	// Identifiers and basic type literals

	IDENT  // main
	INT    // 12345
	FLOAT  // 123.45
	IMAG   // 123.45i
	STRING // "abc"

	LPAREN // (
	RPAREN // )
	COMMA  // ,
	PERIOD // .

	// Operators

	LSS // <
	GTR // >
	LEQ // <=
	GEQ // >=
	EQL // =
	NEQ // <>

	// Keywords

	IN
	IF
	AND
	OR
	NOT
	CASE
	INSERT
	CREATE
	DELETE
	UPDATE
	WITH
)
