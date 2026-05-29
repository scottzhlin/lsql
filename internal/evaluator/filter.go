package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scottlin/lsql/internal/parser"
)

// ApplyFilter returns the subset of rows matching expr. A nil expr returns all rows.
func ApplyFilter(rows []FileRow, expr parser.Expr) ([]FileRow, error) {
	if expr == nil {
		return rows, nil
	}
	var result []FileRow
	for _, row := range rows {
		match, err := evalExpr(expr, row)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: filter error on %q: %v\n", row.Path, err)
			continue
		}
		if match {
			result = append(result, row)
		}
	}
	return result, nil
}

func evalExpr(expr parser.Expr, row FileRow) (bool, error) {
	switch e := expr.(type) {
	case parser.LogicalExpr:
		left, err := evalExpr(e.Left, row)
		if err != nil {
			return false, err
		}
		if e.Op == "AND" && !left {
			return false, nil
		}
		if e.Op == "OR" && left {
			return true, nil
		}
		return evalExpr(e.Right, row)
	case parser.NotExpr:
		v, err := evalExpr(e.Inner, row)
		return !v, err
	case parser.BinaryExpr:
		return evalBinary(e, row)
	case parser.LikeExpr:
		return evalLike(e, row)
	case parser.BetweenExpr:
		return evalBetween(e, row)
	}
	return false, fmt.Errorf("unknown expression type %T", expr)
}

func evalBinary(e parser.BinaryExpr, row FileRow) (bool, error) {
	colVal := row.Get(e.Col)
	if colVal == nil {
		return false, fmt.Errorf("unknown column %q", e.Col)
	}
	return compare(colVal, e.Op, e.Val)
}

func compare(left interface{}, op string, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int64:
		r, err := toInt64(right)
		if err != nil {
			return false, fmt.Errorf("type error: %w", err)
		}
		return compareInt64(l, op, r)
	case string:
		r, ok := right.(string)
		if !ok {
			return false, fmt.Errorf("type mismatch: cannot compare string with %T", right)
		}
		return compareString(l, op, r)
	case bool:
		r, ok := right.(bool)
		if !ok {
			return false, fmt.Errorf("type mismatch: cannot compare bool with %T", right)
		}
		switch op {
		case "=":
			return l == r, nil
		case "!=":
			return l != r, nil
		}
		return false, fmt.Errorf("operator %q not supported for bool", op)
	case time.Time:
		r, ok := right.(time.Time)
		if !ok {
			return false, fmt.Errorf("type mismatch: cannot compare time with %T", right)
		}
		return compareInt64(l.Unix(), op, r.Unix())
	}
	return false, fmt.Errorf("unsupported column type %T", left)
}

func compareInt64(l int64, op string, r int64) (bool, error) {
	switch op {
	case "=":
		return l == r, nil
	case "!=":
		return l != r, nil
	case "<":
		return l < r, nil
	case "<=":
		return l <= r, nil
	case ">":
		return l > r, nil
	case ">=":
		return l >= r, nil
	}
	return false, fmt.Errorf("unknown operator %q", op)
}

func compareString(l, op, r string) (bool, error) {
	switch op {
	case "=":
		return l == r, nil
	case "!=":
		return l != r, nil
	case "<":
		return l < r, nil
	case "<=":
		return l <= r, nil
	case ">":
		return l > r, nil
	case ">=":
		return l >= r, nil
	}
	return false, fmt.Errorf("unknown operator %q for string", op)
}

func toInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	case int:
		return int64(n), nil
	}
	return 0, fmt.Errorf("cannot convert %T to int64", v)
}

func evalLike(e parser.LikeExpr, row FileRow) (bool, error) {
	colVal := row.Get(e.Col)
	s, ok := colVal.(string)
	if !ok {
		return false, fmt.Errorf("LIKE requires a string column, got %T for %q", colVal, e.Col)
	}
	return filepath.Match(likeToGlob(e.Pattern), s)
}

// likeToGlob converts SQL LIKE pattern to filepath.Match glob pattern.
func likeToGlob(pattern string) string {
	// Protect SQL escape sequences first
	pattern = strings.ReplaceAll(pattern, `\%`, "\x00") // escaped % literal
	pattern = strings.ReplaceAll(pattern, `\_`, "\x01") // escaped _ literal
	// Escape filepath.Match special chars that SQL LIKE doesn't use
	pattern = strings.ReplaceAll(pattern, "[", `\[`)
	pattern = strings.ReplaceAll(pattern, "]", `\]`)
	// Translate SQL wildcards to glob wildcards
	pattern = strings.ReplaceAll(pattern, "%", "*")
	pattern = strings.ReplaceAll(pattern, "_", "?")
	// Restore escaped literals
	pattern = strings.ReplaceAll(pattern, "\x00", "%")
	pattern = strings.ReplaceAll(pattern, "\x01", "_")
	return pattern
}

func evalBetween(e parser.BetweenExpr, row FileRow) (bool, error) {
	colVal := row.Get(e.Col)
	low, err := compare(colVal, ">=", e.Low)
	if err != nil {
		return false, err
	}
	if !low {
		return false, nil
	}
	return compare(colVal, "<=", e.High)
}
