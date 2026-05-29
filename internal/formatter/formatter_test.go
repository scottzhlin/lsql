package formatter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/formatter"
)

func makeResult() *evaluator.Result {
	return &evaluator.Result{
		Columns: []string{"name", "size"},
		Rows: []evaluator.ResultRow{
			{"name": "foo.go", "size": int64(1024)},
			{"name": "bar.txt", "size": int64(512)},
		},
	}
}

func TestFormatter_Table(t *testing.T) {
	var sb strings.Builder
	err := formatter.Print(&sb, makeResult(), formatter.Table)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "name") || !strings.Contains(out, "size") {
		t.Error("table output should contain column headers")
	}
	if !strings.Contains(out, "foo.go") || !strings.Contains(out, "1024") {
		t.Error("table output should contain row data")
	}
}

func TestFormatter_CSV(t *testing.T) {
	var sb strings.Builder
	err := formatter.Print(&sb, makeResult(), formatter.CSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(sb.String()), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Fatalf("expected 3 CSV lines, got %d: %q", len(lines), sb.String())
	}
	if lines[0] != "name,size" {
		t.Errorf("CSV header: got %q, want %q", lines[0], "name,size")
	}
}

func TestFormatter_JSON(t *testing.T) {
	var sb strings.Builder
	err := formatter.Print(&sb, makeResult(), formatter.JSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, `"foo.go"`) {
		t.Error("JSON output should contain foo.go")
	}
	if !strings.Contains(out, `"name"`) {
		t.Error("JSON output should contain key 'name'")
	}
}

func TestFormatter_TimeFormatting(t *testing.T) {
	result := &evaluator.Result{
		Columns: []string{"modified"},
		Rows: []evaluator.ResultRow{
			{"modified": time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		},
	}
	var sb strings.Builder
	formatter.Print(&sb, result, formatter.Table)
	out := sb.String()
	if !strings.Contains(out, "2024-01-15") {
		t.Errorf("time should be formatted as YYYY-MM-DD HH:MM:SS, got: %q", out)
	}
}

func TestFormatter_UnknownFormat(t *testing.T) {
	var sb strings.Builder
	err := formatter.Print(&sb, makeResult(), "xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}
