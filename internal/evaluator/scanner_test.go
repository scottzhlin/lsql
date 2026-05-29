package evaluator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/parser"
)

func TestScanner_Flat(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), "hello")
	writeFile(t, filepath.Join(dir, "b.txt"), "world")
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	rows, err := evaluator.Scan(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Depth != 0 {
			t.Errorf("expected depth 0 for flat scan, got %d", r.Depth)
		}
	}
}

func TestScanner_Recursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "root.go"), "")
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	writeFile(t, filepath.Join(sub, "child.go"), "")

	rows, err := evaluator.Scan(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 3 { // root.go, sub/, child.go
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	depths := map[string]int{}
	for _, r := range rows {
		depths[r.Name] = r.Depth
	}
	if depths["root.go"] != 0 {
		t.Errorf("root.go: expected depth 0, got %d", depths["root.go"])
	}
	if depths["sub"] != 0 {
		t.Errorf("sub: expected depth 0, got %d", depths["sub"])
	}
	if depths["child.go"] != 1 {
		t.Errorf("child.go: expected depth 1, got %d", depths["child.go"])
	}
}

func TestScanner_FileRowFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hello.go"), "package main")

	rows, err := evaluator.Scan(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Name != "hello.go" {
		t.Errorf("Name: got %q, want %q", r.Name, "hello.go")
	}
	if r.Extension != ".go" {
		t.Errorf("Extension: got %q, want %q", r.Extension, ".go")
	}
	if r.IsDir {
		t.Error("expected IsDir=false")
	}
	if r.Size != int64(len("package main")) {
		t.Errorf("Size: got %d, want %d", r.Size, len("package main"))
	}
	if r.Path == "" {
		t.Error("Path should not be empty")
	}
}

func TestScanner_InvalidPath(t *testing.T) {
	_, err := evaluator.Scan("/nonexistent/path/xyz", false)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func TestFilter_BinaryExpr(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "small.go"), "hi")
	writeFile(t, filepath.Join(dir, "big.go"), strings.Repeat("x", 1000))

	rows, _ := evaluator.Scan(dir, false)
	filtered, err := evaluator.ApplyFilter(rows, parser.BinaryExpr{Col: "size", Op: ">", Val: int64(100)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "big.go" {
		t.Errorf("expected [big.go], got %v", names(filtered))
	}
}

func TestFilter_LikeExpr(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.go"), "")
	writeFile(t, filepath.Join(dir, "main.txt"), "")

	rows, _ := evaluator.Scan(dir, false)
	filtered, err := evaluator.ApplyFilter(rows, parser.LikeExpr{Col: "name", Pattern: "%.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "main.go" {
		t.Errorf("expected [main.go], got %v", names(filtered))
	}
}

func TestFilter_LogicalAnd(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.go"), strings.Repeat("x", 500))
	writeFile(t, filepath.Join(dir, "b.go"), strings.Repeat("x", 2000))
	writeFile(t, filepath.Join(dir, "c.txt"), strings.Repeat("x", 2000))

	rows, _ := evaluator.Scan(dir, false)
	expr := parser.LogicalExpr{
		Left:  parser.BinaryExpr{Col: "extension", Op: "=", Val: ".go"},
		Op:    "AND",
		Right: parser.BinaryExpr{Col: "size", Op: ">", Val: int64(1000)},
	}
	filtered, err := evaluator.ApplyFilter(rows, expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "b.go" {
		t.Errorf("expected [b.go], got %v", names(filtered))
	}
}

func TestFilter_NilExpr(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "file.go"), "")
	rows, _ := evaluator.Scan(dir, false)
	filtered, err := evaluator.ApplyFilter(rows, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != len(rows) {
		t.Errorf("nil filter should return all rows")
	}
}

func names(rows []evaluator.FileRow) []string {
	ns := make([]string, len(rows))
	for i, r := range rows {
		ns[i] = r.Name
	}
	return ns
}
