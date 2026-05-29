package formatter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/scottlin/lsql/internal/evaluator"
)

type Format string

const (
	Table Format = "table"
	CSV   Format = "csv"
	JSON  Format = "json"
)

func Print(w io.Writer, result *evaluator.Result, format Format) error {
	switch format {
	case Table:
		return printTable(w, result)
	case CSV:
		return printCSV(w, result)
	case JSON:
		return printJSON(w, result)
	}
	return fmt.Errorf("unknown format %q (valid: table, csv, json)", format)
}

func printTable(w io.Writer, result *evaluator.Result) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(result.Columns, "\t"))
	for _, row := range result.Rows {
		vals := rowValues(result.Columns, row)
		fmt.Fprintln(tw, strings.Join(vals, "\t"))
	}
	return tw.Flush()
}

func printCSV(w io.Writer, result *evaluator.Result) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(result.Columns); err != nil {
		return err
	}
	for _, row := range result.Rows {
		if err := cw.Write(rowValues(result.Columns, row)); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func printJSON(w io.Writer, result *evaluator.Result) error {
	out := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		m := make(map[string]interface{}, len(result.Columns))
		for _, col := range result.Columns {
			m[col] = formatValue(row[col])
		}
		out[i] = m
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func rowValues(columns []string, row evaluator.ResultRow) []string {
	vals := make([]string, len(columns))
	for i, col := range columns {
		vals[i] = fmt.Sprintf("%v", formatValue(row[col]))
	}
	return vals
}

func formatValue(v interface{}) interface{} {
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return v
}
