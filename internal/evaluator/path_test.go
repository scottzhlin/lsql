package evaluator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/lexer"
	"github.com/scottzhlin/lsql/internal/parser"
)

func TestResolvePath_Home(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	result := runSQL(t, "SELECT name FROM "+dir)
	if result.Meta.From != dir {
		t.Fatalf("expected FROM %q, got %q", dir, result.Meta.From)
	}
	_ = home
}

func TestRun_MetaPopulated(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "x")
	writeFile(t, filepath.Join(dir, "b.txt"), "y")

	result := runSQL(t, "SELECT name FROM "+dir+" WHERE extension = '.go'")
	if result.Meta.Scanned != 2 {
		t.Errorf("Scanned: got %d want 2", result.Meta.Scanned)
	}
	if result.Meta.Matched != 1 {
		t.Errorf("Matched: got %d want 1", result.Meta.Matched)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("rows: got %d", len(result.Rows))
	}
	if result.Meta.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRun_TildeFrom(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	tokens, err := lexer.New("SELECT name FROM ~ LIMIT 1").Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	result, err := evaluator.Run(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Meta.From != home {
		t.Errorf("FROM ~ should resolve to home, got %q", result.Meta.From)
	}
}
