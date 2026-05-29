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
		Meta: evaluator.Meta{Scanned: 2, Matched: 2, Duration: 3 * time.Millisecond},
	}
}

func TestFormatter_Table(t *testing.T) {
	var out, summary strings.Builder
	opts := formatter.DefaultOptions(formatter.Table)
	err := formatter.Print(&out, &summary, makeResult(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := out.String()
	if !strings.Contains(body, "name") || !strings.Contains(body, "foo.go") {
		t.Errorf("table output missing data: %q", body)
	}
	if !strings.Contains(summary.String(), "2 rows") {
		t.Errorf("expected summary with row count, got %q", summary.String())
	}
}

func TestFormatter_TableEmptyNoHeaders(t *testing.T) {
	result := &evaluator.Result{
		Columns: []string{"name"},
		Rows:    nil,
		Meta:    evaluator.Meta{Scanned: 5, Duration: time.Millisecond},
	}
	var out, summary strings.Builder
	opts := formatter.DefaultOptions(formatter.Table)
	if err := formatter.Print(&out, &summary, result, opts); err != nil {
		t.Fatal(err)
	}
	if out.String() != "" {
		t.Errorf("empty table should produce no stdout, got %q", out.String())
	}
	if !strings.Contains(summary.String(), "(empty)") {
		t.Errorf("summary should mark empty result: %q", summary.String())
	}
}

func TestFormatter_CSV(t *testing.T) {
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.CSV)
	opts.Quiet = true
	err := formatter.Print(&out, nil, makeResult(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 CSV lines, got %d: %q", len(lines), out.String())
	}
	if lines[0] != "name,size" {
		t.Errorf("CSV header: got %q, want %q", lines[0], "name,size")
	}
}

func TestFormatter_CSVEmpty(t *testing.T) {
	result := &evaluator.Result{Columns: []string{"name"}, Meta: evaluator.Meta{Scanned: 1}}
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.CSV)
	opts.Quiet = true
	if err := formatter.Print(&out, nil, result, opts); err != nil {
		t.Fatal(err)
	}
	if out.String() != "" {
		t.Errorf("empty csv with headers=auto should be silent, got %q", out.String())
	}
}

func TestFormatter_JSON(t *testing.T) {
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.JSON)
	opts.Quiet = true
	err := formatter.Print(&out, nil, makeResult(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"foo.go"`) {
		t.Error("JSON output should contain foo.go")
	}
}

func TestFormatter_TimeFormatting(t *testing.T) {
	result := &evaluator.Result{
		Columns: []string{"modified"},
		Rows: []evaluator.ResultRow{
			{"modified": time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		},
		Meta: evaluator.Meta{Scanned: 1, Matched: 1},
	}
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.Table)
	opts.Quiet = true
	formatter.Print(&out, nil, result, opts)
	if !strings.Contains(out.String(), "2024-01-15") {
		t.Errorf("time should be formatted, got: %q", out.String())
	}
}

func TestFormatter_HumanSize(t *testing.T) {
	result := &evaluator.Result{
		Columns: []string{"size"},
		Rows:    []evaluator.ResultRow{{"size": int64(1536)}},
		Meta:    evaluator.Meta{Scanned: 1, Matched: 1},
	}
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.Table)
	opts.Quiet = true
	formatter.Print(&out, nil, result, opts)
	if !strings.Contains(out.String(), "KiB") {
		t.Errorf("human size expected KiB, got %q", out.String())
	}
}

func TestFormatter_UnknownFormat(t *testing.T) {
	var out strings.Builder
	err := formatter.Print(&out, nil, makeResult(), formatter.Options{Format: "xml"})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestParseHeaderMode(t *testing.T) {
	if _, err := formatter.ParseHeaderMode("bogus"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatter_JSONMeta(t *testing.T) {
	var out strings.Builder
	opts := formatter.DefaultOptions(formatter.JSON)
	opts.Quiet = true
	opts.JSONMeta = true
	if err := formatter.Print(&out, nil, makeResult(), opts); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	for _, want := range []string{`"meta"`, `"data"`, `"row_count"`, `"scanned"`, `"foo.go"`} {
		if !strings.Contains(body, want) {
			t.Errorf("json-meta output missing %q in %q", want, body)
		}
	}
}
