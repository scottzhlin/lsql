package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/parser"
)

func makeResultRows(sizes ...int64) []evaluator.ResultRow {
	rows := make([]evaluator.ResultRow, len(sizes))
	for i, s := range sizes {
		rows[i] = evaluator.ResultRow{"size": s, "name": fmt.Sprintf("file%d", i)}
	}
	return rows
}

func TestSorter_OrderByAsc(t *testing.T) {
	rows := makeResultRows(300, 100, 200)
	sorted := evaluator.SortAndLimit(rows, []parser.OrderItem{{Col: "size"}}, -1)
	sizes := make([]int64, len(sorted))
	for i, r := range sorted {
		sizes[i] = r["size"].(int64)
	}
	want := []int64{100, 200, 300}
	for i := range want {
		if sizes[i] != want[i] {
			t.Errorf("sorted[%d]: got %d, want %d", i, sizes[i], want[i])
		}
	}
}

func TestSorter_OrderByDesc(t *testing.T) {
	rows := makeResultRows(100, 300, 200)
	sorted := evaluator.SortAndLimit(rows, []parser.OrderItem{{Col: "size", Desc: true}}, -1)
	if sorted[0]["size"].(int64) != 300 {
		t.Errorf("first row should be 300, got %d", sorted[0]["size"])
	}
}

func TestSorter_Limit(t *testing.T) {
	rows := makeResultRows(1, 2, 3, 4, 5)
	sorted := evaluator.SortAndLimit(rows, nil, 3)
	if len(sorted) != 3 {
		t.Errorf("expected 3 rows after LIMIT 3, got %d", len(sorted))
	}
}

func TestSorter_NoOp(t *testing.T) {
	rows := makeResultRows(5, 3, 1)
	sorted := evaluator.SortAndLimit(rows, nil, -1)
	if len(sorted) != 3 {
		t.Errorf("expected 3 rows, got %d", len(sorted))
	}
	if sorted[0]["size"].(int64) != 5 {
		t.Errorf("expected first row size=5, got %v", sorted[0]["size"])
	}
}
