package formatter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/scottzhlin/lsql/internal/evaluator"
)

type Format string

const (
	Table Format = "table"
	CSV   Format = "csv"
	JSON  Format = "json"
)

// HeaderMode controls when column headers are emitted.
type HeaderMode string

const (
	HeadersAuto   HeaderMode = "auto"   // only when there is at least one row
	HeadersAlways HeaderMode = "always"
	HeadersNever  HeaderMode = "never"
)

// Options configures query output.
type Options struct {
	Format    Format
	Headers   HeaderMode
	HumanSize bool
	// Quiet suppresses the stderr summary line.
	Quiet bool
}

// DefaultOptions is the standard interactive CLI configuration.
func DefaultOptions(format Format) Options {
	return Options{
		Format:    format,
		Headers:   HeadersAuto,
		HumanSize: true,
	}
}

// Print writes query results to out and an optional summary to summary (typically stderr).
func Print(out, summary io.Writer, result *evaluator.Result, opts Options) error {
	if opts.Headers == "" {
		opts.Headers = HeadersAuto
	}
	showHeaders := shouldShowHeaders(opts.Headers, len(result.Rows))

	switch opts.Format {
	case Table:
		if err := printTable(out, result, showHeaders, opts.HumanSize); err != nil {
			return err
		}
	case CSV:
		if err := printCSV(out, result, showHeaders, opts.HumanSize); err != nil {
			return err
		}
	case JSON:
		if err := printJSON(out, result, opts.HumanSize); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q (valid: table, csv, json)", opts.Format)
	}

	if !opts.Quiet && summary != nil {
		fmt.Fprintln(summary, FormatSummary(result))
	}
	return nil
}

// FormatSummary renders a one-line execution summary for stderr.
func FormatSummary(result *evaluator.Result) string {
	n := len(result.Rows)
	parts := []string{formatCount(n, "row", "rows"), formatDuration(result.Meta.Duration)}

	if result.Meta.Scanned > 0 {
		parts = append(parts, fmt.Sprintf("scanned %s", formatCount(result.Meta.Scanned, "entry", "entries")))
	}
	if result.Meta.Matched != n && result.Meta.Matched != result.Meta.Scanned {
		parts = append(parts, fmt.Sprintf("%s matched filter", formatCount(result.Meta.Matched, "row", "rows")))
	}
	if result.Meta.Warnings > 0 {
		parts = append(parts, fmt.Sprintf("%s skipped", formatCount(result.Meta.Warnings, "path", "paths")))
	}
	if n == 0 {
		return "→ (empty) · " + strings.Join(parts, " · ")
	}
	return "→ " + strings.Join(parts, " · ")
}

func shouldShowHeaders(mode HeaderMode, rowCount int) bool {
	switch mode {
	case HeadersNever:
		return false
	case HeadersAlways:
		return true
	default:
		return rowCount > 0
	}
}

func printTable(w io.Writer, result *evaluator.Result, showHeaders, humanSize bool) error {
	if len(result.Rows) == 0 {
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if showHeaders {
		fmt.Fprintln(tw, strings.Join(result.Columns, "\t"))
	}
	for _, row := range result.Rows {
		vals := rowValues(result.Columns, row, humanSize)
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}
	return tw.Flush()
}

func printCSV(w io.Writer, result *evaluator.Result, showHeaders, humanSize bool) error {
	if len(result.Rows) == 0 && !showHeaders {
		return nil
	}
	cw := csv.NewWriter(w)
	if showHeaders {
		if err := cw.Write(result.Columns); err != nil {
			return err
		}
	}
	for _, row := range result.Rows {
		if err := cw.Write(rowValues(result.Columns, row, humanSize)); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func printJSON(w io.Writer, result *evaluator.Result, humanSize bool) error {
	out := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		m := make(map[string]interface{}, len(result.Columns))
		for _, col := range result.Columns {
			m[col] = formatCell(col, row[col], humanSize)
		}
		out[i] = m
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func rowValues(columns []string, row evaluator.ResultRow, humanSize bool) []string {
	vals := make([]string, len(columns))
	for i, col := range columns {
		vals[i] = fmt.Sprintf("%v", formatCell(col, row[col], humanSize))
	}
	return vals
}

func formatCell(col string, v interface{}, humanSize bool) interface{} {
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	if humanSize && isSizeColumn(col) {
		if n, ok := v.(int64); ok {
			return formatBytes(n)
		}
	}
	return v
}

func isSizeColumn(col string) bool {
	c := strings.ToLower(col)
	return c == "size" || strings.HasSuffix(c, "(size)") || c == "sum(size)" || c == "min(size)" || c == "max(size)"
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit && exp < 3; n /= unit {
		div *= unit
		exp++
	}
	val := float64(b) / float64(div)
	suffix := []string{"KiB", "MiB", "GiB", "TiB"}[exp]
	if val >= 100 {
		return fmt.Sprintf("%.0f %s", val, suffix)
	}
	if val == float64(int64(val)) {
		return fmt.Sprintf("%.0f %s", val, suffix)
	}
	return fmt.Sprintf("%.1f %s", val, suffix)
}

func formatCount(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// ParseHeaderMode parses a CLI --headers value.
func ParseHeaderMode(s string) (HeaderMode, error) {
	switch HeaderMode(s) {
	case HeadersAuto, HeadersAlways, HeadersNever, "":
		if s == "" {
			return HeadersAuto, nil
		}
		return HeaderMode(s), nil
	default:
		return "", fmt.Errorf("invalid headers mode %q (valid: auto, always, never)", s)
	}
}
