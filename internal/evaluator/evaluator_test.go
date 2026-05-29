package evaluator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/lexer"
	"github.com/scottzhlin/lsql/internal/parser"
)

func runSQL(t *testing.T, sql string) *evaluator.Result {
	t.Helper()
	tokens, err := lexer.New(sql).Tokenize()
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	stmt, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := evaluator.Run(stmt)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return result
}

func TestEvaluator_SelectStar(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.go"), "package main")
	writeFile(t, filepath.Join(dir, "readme.txt"), "hello")

	result := runSQL(t, "SELECT * FROM "+dir)
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
	if len(result.Columns) != 8 {
		t.Fatalf("SELECT * should produce 8 columns, got %d", len(result.Columns))
	}
}

func TestEvaluator_WhereExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "")
	writeFile(t, filepath.Join(dir, "b.go"), "")
	writeFile(t, filepath.Join(dir, "c.txt"), "")

	result := runSQL(t, "SELECT name FROM "+dir+" WHERE extension = '.go'")
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 .go files, got %d", len(result.Rows))
	}
}

func TestEvaluator_OrderByLimit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "x")
	writeFile(t, filepath.Join(dir, "b.go"), "xxx")
	writeFile(t, filepath.Join(dir, "c.go"), "xx")

	result := runSQL(t, "SELECT name, size FROM "+dir+" ORDER BY size DESC LIMIT 2")
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "b.go" {
		t.Errorf("first row should be b.go (largest), got %v", result.Rows[0]["name"])
	}
}

func TestEvaluator_Recursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "root.go"), "")
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	writeFile(t, filepath.Join(sub, "child.go"), "")

	result := runSQL(t, "SELECT name FROM "+dir+" RECURSIVE WHERE extension = '.go'")
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 .go files recursively, got %d", len(result.Rows))
	}
}

func TestEvaluator_GroupByCount(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "")
	writeFile(t, filepath.Join(dir, "b.go"), "")
	writeFile(t, filepath.Join(dir, "c.txt"), "")

	result := runSQL(t, "SELECT extension, COUNT(*) FROM "+dir+" GROUP BY extension ORDER BY extension ASC")
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result.Rows))
	}
	found := false
	for _, r := range result.Rows {
		if r["extension"] == ".go" && r["COUNT(*)"] == int64(2) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected .go group with COUNT(*)=2, rows: %v", result.Rows)
	}
}
