package evaluator

import (
	"github.com/scottzhlin/lsql/internal/parser"
)

// Result is the complete output of a query.
type Result struct {
	Columns []string
	Rows    []ResultRow
}

// Run executes a parsed SELECT statement against the filesystem.
func Run(stmt *parser.SelectStmt) (*Result, error) {
	rows, err := Scan(stmt.From, stmt.Recursive)
	if err != nil {
		return nil, err
	}

	rows, err = ApplyFilter(rows, stmt.Where)
	if err != nil {
		return nil, err
	}

	resultRows, err := Project(rows, stmt)
	if err != nil {
		return nil, err
	}

	resultRows = SortAndLimit(resultRows, stmt.OrderBy, stmt.Limit)

	return &Result{
		Columns: columnNames(stmt),
		Rows:    resultRows,
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
