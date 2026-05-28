package evaluator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scottlin/lsql/internal/evaluator"
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
