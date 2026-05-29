package evaluator

import (
	"time"

	"github.com/scottzhlin/lsql/internal/parser"
)

// RunOptions configures query execution (warnings, etc.).
type RunOptions struct {
	// Verbose prints filesystem warnings to stderr as they occur.
	Verbose bool
}

// Result is the complete output of a query.
type Result struct {
	Columns []string
	Rows    []ResultRow
	Meta    Meta
}

// Run executes a parsed SELECT statement against the filesystem.
func Run(stmt *parser.SelectStmt) (*Result, error) {
	return RunWithOptions(stmt, RunOptions{})
}

// RunWithOptions executes a query with optional warning verbosity.
func RunWithOptions(stmt *parser.SelectStmt, opts RunOptions) (*Result, error) {
	start := time.Now()
	st := &scanStats{verbose: opts.Verbose}

	from, err := resolvePath(stmt.From)
	if err != nil {
		return nil, err
	}

	rows, err := Scan(from, stmt.Recursive, st)
	if err != nil {
		return nil, err
	}
	scanned := len(rows)

	rows, err = ApplyFilter(rows, stmt.Where, st)
	if err != nil {
		return nil, err
	}
	matched := len(rows)

	resultRows, err := Project(rows, stmt)
	if err != nil {
		return nil, err
	}

	resultRows = SortAndLimit(resultRows, stmt.OrderBy, stmt.Limit)

	meta := Meta{
		From:      from,
		Recursive: stmt.Recursive,
		Scanned:   scanned,
		Matched:   matched,
		Duration:  time.Since(start),
		Warnings:  st.warns,
	}

	return &Result{
		Columns: columnNames(stmt),
		Rows:    resultRows,
		Meta:    meta,
	}, nil
}

func columnNames(stmt *parser.SelectStmt) []string {
	if len(stmt.Columns) == 1 && stmt.Columns[0].Col == "*" {
		return []string{"name", "path", "size", "mode", "modified", "is_dir", "extension", "depth"}
	}
	names := make([]string, len(stmt.Columns))
	for i, c := range stmt.Columns {
		names[i] = c.String()
	}
	return names
}
