package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

var keywords = map[string]TokenType{
	"SELECT":    TOKEN_SELECT,
	"FROM":      TOKEN_FROM,
	"WHERE":     TOKEN_WHERE,
	"AND":       TOKEN_AND,
	"OR":        TOKEN_OR,
	"NOT":       TOKEN_NOT,
	"LIKE":      TOKEN_LIKE,
	"BETWEEN":   TOKEN_BETWEEN,
	"GROUP":     TOKEN_GROUP,
	"BY":        TOKEN_BY,
	"ORDER":     TOKEN_ORDER,
	"ASC":       TOKEN_ASC,
	"DESC":      TOKEN_DESC,
	"LIMIT":     TOKEN_LIMIT,
	"RECURSIVE": TOKEN_RECURSIVE,
	"COUNT":     TOKEN_COUNT,
	"SUM":       TOKEN_SUM,
	"MIN":       TOKEN_MIN,
	"MAX":       TOKEN_MAX,
	"TRUE":      TOKEN_BOOL,
	"FALSE":     TOKEN_BOOL,
}

type Lexer struct {
	input []rune
	pos   int
}

func New(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens, nil
}

func (l *Lexer) nextToken() (Token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return Token{Type: TOKEN_EOF, Col: l.pos}, nil
	}

	col := l.pos
	ch := l.input[l.pos]

	switch {
	case ch == '\'':
		s, err := l.readString(col)
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TOKEN_STRING, Literal: s, Col: col}, nil

	case unicode.IsDigit(ch):
		return Token{Type: TOKEN_NUMBER, Literal: l.readNumber(), Col: col}, nil

	case unicode.IsLetter(ch) || ch == '_' || ch == '.' || ch == '/' || ch == '~':
		word := l.readWord()
		if tt, ok := keywords[strings.ToUpper(word)]; ok {
			return Token{Type: tt, Literal: word, Col: col}, nil
		}
		return Token{Type: TOKEN_IDENT, Literal: word, Col: col}, nil

	case ch == '=':
		l.pos++
		return Token{Type: TOKEN_EQ, Literal: "=", Col: col}, nil

	case ch == '!':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2
			return Token{Type: TOKEN_NEQ, Literal: "!=", Col: col}, nil
		}
		l.pos++
		return Token{}, fmt.Errorf("lexer error at col %d: unexpected character '!'", col+1)

	case ch == '<':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2
			return Token{Type: TOKEN_LTE, Literal: "<=", Col: col}, nil
		}
		l.pos++
		return Token{Type: TOKEN_LT, Literal: "<", Col: col}, nil

	case ch == '>':
		if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
			l.pos += 2
			return Token{Type: TOKEN_GTE, Literal: ">=", Col: col}, nil
		}
		l.pos++
		return Token{Type: TOKEN_GT, Literal: ">", Col: col}, nil

	case ch == '*':
		l.pos++
		return Token{Type: TOKEN_STAR, Literal: "*", Col: col}, nil

	case ch == '(':
		l.pos++
		return Token{Type: TOKEN_LPAREN, Literal: "(", Col: col}, nil

	case ch == ')':
		l.pos++
		return Token{Type: TOKEN_RPAREN, Literal: ")", Col: col}, nil

	case ch == ',':
		l.pos++
		return Token{Type: TOKEN_COMMA, Literal: ",", Col: col}, nil
	}

	l.pos++
	return Token{}, fmt.Errorf("lexer error at col %d: unexpected character %q", col+1, ch)
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readString(col int) (string, error) {
	l.pos++ // skip opening '
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		l.pos++
		if ch == '\'' {
			return sb.String(), nil
		}
		sb.WriteRune(ch)
	}
	return "", fmt.Errorf("lexer error at col %d: unterminated string literal", col+1)
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readWord() string {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' || ch == '/' || ch == '~' || ch == '-' {
			l.pos++
		} else {
			break
		}
	}
	return string(l.input[start:l.pos])
}
