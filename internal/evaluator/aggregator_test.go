package evaluator_test

import (
	"testing"
	"time"

	"github.com/scottlin/lsql/internal/evaluator"
	"github.com/scottlin/lsql/internal/parser"
)

func makeRows(sizes ...int64) []evaluator.FileRow {
	rows := make([]evaluator.FileRow, len(sizes))
	for i, s := range sizes {
		rows[i] = evaluator.FileRow{
			Name:      "file.go",
			Extension: ".go",
			Size:      s,
			Modified:  time.Now(),
		}
	}
	return rows
}

func TestProject_PlainColumns(t *testing.T) {
	rows := makeRows(100, 200)
	stmt := &parser.SelectStmt{
		Columns: []parser.ColExpr{{Col: "size"}},
	}
	result, err := evaluator.Project(rows, stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 result rows, got %d", len(result))
	}
	if result[0]["size"] != int64(100) {
		t.Errorf("expected size=100, got %v", result[0]["size"])
	}
}

func TestProject_SelectStar(t *testing.T) {
	rows := makeRows(42)
	stmt := &parser.SelectStmt{
		Columns: []parser.ColExpr{{Col: "*"}},
	}
	result, err := evaluator.Project(rows, stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result[0]["name"]; !ok {
		t.Error("SELECT * should include 'name' column")
	}
	if _, ok := result[0]["size"]; !ok {
		t.Error("SELECT * should include 'size' column")
	}
}

func TestProject_Count(t *testing.T) {
	rows := makeRows(1, 2, 3)
	stmt := &parser.SelectStmt{
		Columns: []parser.ColExpr{{Agg: "COUNT", Col: "*"}},
	}
	result, err := evaluator.Project(rows, stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 aggregate row, got %d", len(result))
	}
	if result[0]["COUNT(*)"] != int64(3) {
		t.Errorf("expected COUNT(*)=3, got %v", result[0]["COUNT(*)"])
	}
}

func TestProject_SumAndGroupBy(t *testing.T) {
	rows := []evaluator.FileRow{
		{Name: "a.go", Extension: ".go", Size: 100},
		{Name: "b.go", Extension: ".go", Size: 200},
		{Name: "c.txt", Extension: ".txt", Size: 50},
	}
	stmt := &parser.SelectStmt{
		Columns: []parser.ColExpr{
			{Col: "extension"},
			{Agg: "SUM", Col: "size"},
		},
		GroupBy: []string{"extension"},
	}
	result, err := evaluator.Project(rows, stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result))
	}
	totals := map[string]int64{}
	for _, r := range result {
		ext := r["extension"].(string)
		totals[ext] = r["SUM(size)"].(int64)
	}
	if totals[".go"] != 300 {
		t.Errorf(".go total: got %d, want 300", totals[".go"])
	}
	if totals[".txt"] != 50 {
		t.Errorf(".txt total: got %d, want 50", totals[".txt"])
	}
}
