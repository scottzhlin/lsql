package lexer

type TokenType int

const (
	TOKEN_IDENT  TokenType = iota // identifier or path
	TOKEN_STRING                  // 'value'
	TOKEN_NUMBER                  // 42, 1024
	TOKEN_BOOL                    // true, false

	TOKEN_SELECT
	TOKEN_FROM
	TOKEN_WHERE
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_LIKE
	TOKEN_BETWEEN
	TOKEN_GROUP
	TOKEN_BY
	TOKEN_ORDER
	TOKEN_ASC
	TOKEN_DESC
	TOKEN_LIMIT
	TOKEN_RECURSIVE
	TOKEN_COUNT
	TOKEN_SUM
	TOKEN_MIN
	TOKEN_MAX

	TOKEN_EQ   // =
	TOKEN_NEQ  // !=
	TOKEN_LT   // <
	TOKEN_LTE  // <=
	TOKEN_GT   // >
	TOKEN_GTE  // >=
	TOKEN_STAR // *

	TOKEN_LPAREN // (
	TOKEN_RPAREN // )
	TOKEN_COMMA  // ,

	TOKEN_EOF
)

type Token struct {
	Type    TokenType
	Literal string
	Col     int
}
