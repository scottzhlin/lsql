package lexer_test

import (
	"testing"

	"github.com/scottlin/lsql/internal/lexer"
)

func tokenTypes(tokens []lexer.Token) []lexer.TokenType {
	types := make([]lexer.TokenType, len(tokens))
	for i, t := range tokens {
		types[i] = t.Type
	}
	return types
}

func TestLexer_SimpleSelect(t *testing.T) {
	tokens, err := lexer.New("SELECT name FROM .").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []lexer.TokenType{
		lexer.TOKEN_SELECT,
		lexer.TOKEN_IDENT, // name
		lexer.TOKEN_FROM,
		lexer.TOKEN_IDENT, // .
		lexer.TOKEN_EOF,
	}
	got := tokenTypes(tokens)
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestLexer_Operators(t *testing.T) {
	tokens, err := lexer.New("size >= 1000 AND name != 'foo'").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []lexer.TokenType{
		lexer.TOKEN_IDENT,  // size
		lexer.TOKEN_GTE,    // >=
		lexer.TOKEN_NUMBER, // 1000
		lexer.TOKEN_AND,
		lexer.TOKEN_IDENT,  // name
		lexer.TOKEN_NEQ,    // !=
		lexer.TOKEN_STRING, // foo
		lexer.TOKEN_EOF,
	}
	got := tokenTypes(tokens)
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestLexer_AggregateAndRecursive(t *testing.T) {
	tokens, err := lexer.New("SELECT COUNT(*), SUM(size) FROM . RECURSIVE").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []lexer.TokenType{
		lexer.TOKEN_SELECT,
		lexer.TOKEN_COUNT,
		lexer.TOKEN_LPAREN,
		lexer.TOKEN_STAR,
		lexer.TOKEN_RPAREN,
		lexer.TOKEN_COMMA,
		lexer.TOKEN_SUM,
		lexer.TOKEN_LPAREN,
		lexer.TOKEN_IDENT, // size
		lexer.TOKEN_RPAREN,
		lexer.TOKEN_FROM,
		lexer.TOKEN_IDENT, // .
		lexer.TOKEN_RECURSIVE,
		lexer.TOKEN_EOF,
	}
	got := tokenTypes(tokens)
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestLexer_BoolAndLike(t *testing.T) {
	tokens, err := lexer.New("is_dir = true AND name LIKE '%.go'").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []lexer.TokenType{
		lexer.TOKEN_IDENT,  // is_dir
		lexer.TOKEN_EQ,
		lexer.TOKEN_BOOL,   // true
		lexer.TOKEN_AND,
		lexer.TOKEN_IDENT,  // name
		lexer.TOKEN_LIKE,
		lexer.TOKEN_STRING, // %.go
		lexer.TOKEN_EOF,
	}
	got := tokenTypes(tokens)
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestLexer_IllegalChar(t *testing.T) {
	_, err := lexer.New("SELECT # FROM .").Tokenize()
	if err == nil {
		t.Fatal("expected error for illegal character, got nil")
	}
}

func TestLexer_UnterminatedString(t *testing.T) {
	_, err := lexer.New("name = 'unterminated").Tokenize()
	if err == nil {
		t.Fatal("expected error for unterminated string, got nil")
	}
}

func TestLexer_Path(t *testing.T) {
	tokens, err := lexer.New("SELECT name FROM /usr/local/bin").Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[3].Literal != "/usr/local/bin" {
		t.Errorf("path token: got %q, want %q", tokens[3].Literal, "/usr/local/bin")
	}
}
