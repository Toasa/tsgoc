package token

// TokenType is assigned unique number
type TokenType string

// tokentype
const (
	INT       = "INT"       // 20, 255
	ADD       = "ADD"       // '+'
	SUB       = "SUB"       // '-'
	MUL       = "MUL"       // '*'
	DIV       = "DIV"       // '/'
	REM       = "REM"       // '%'
	LPAREN    = "LPAREN"    // '('
	RPAREN    = "RPAREN"    // ')'
	IDENT     = "IDENT"     // abc, toasa
	SVDECL    = "SVDECL"    // ':=' Short Var Declaration
	SEMICOLON = "SEMICOLON" // ';'
	EOF       = "EOF"       // End of file
)

// Token (minimum unit of Go code)
type Token struct {
	Type    TokenType
	Literal string
}

// New token
func New(t TokenType, lit string) Token {
	return Token{Type: t, Literal: lit}
}
