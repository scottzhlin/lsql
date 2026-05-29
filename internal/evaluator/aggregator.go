package evaluator

import (
	"fmt"
	"strings"
	"time"

	"github.com/scottlin/lsql/internal/parser"
)

// ResultRow is one row in the query output.
type ResultRow map[string]interface{}

// Project applies SELECT (with optional aggregation) to produce ResultRows.
func Project(rows []FileRow, stmt *parser.SelectStmt) ([]ResultRow, error) {
	if len(stmt.GroupBy) > 0 || hasAggregates(stmt.Columns) {
		return aggregate(rows, stmt)
	}
	return projectPlain(rows, stmt.Columns)
}

func hasAggregates(cols []parser.ColExpr) bool {
	for _, c := range cols {
		if c.Agg != "" {
			return true
		}
	}
	return false
}

func projectPlain(rows []FileRow, cols []parser.ColExpr) ([]ResultRow, error) {
	result := make([]ResultRow, 0, len(rows))
	for _, row := range rows {
		r, err := projectRow(row, cols)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func projectRow(row FileRow, cols []parser.ColExpr) (ResultRow, error) {
	if len(cols) == 1 && cols[0].Col == "*" {
		return ResultRow{
			"name":      row.Name,
			"path":      row.Path,
			"size":      row.Size,
			"mode":      row.Mode,
			"modified":  row.Modified,
			"is_dir":    row.IsDir,
			"extension": row.Extension,
			"depth":     int64(row.Depth),
		}, nil
	}
	r := make(ResultRow, len(cols))
	for _, c := range cols {
		if c.Agg != "" {
			return nil, fmt.Errorf("aggregate %s(%s) used without GROUP BY", c.Agg, c.Col)
		}
		val := row.Get(c.Col)
		if val == nil {
			return nil, fmt.Errorf("unknown column %q", c.Col)
		}
		r[c.String()] = val
	}
	return r, nil
}

func aggregate(rows []FileRow, stmt *parser.SelectStmt) ([]ResultRow, error) {
	type groupKey = string
	var order []groupKey
	groups := map[groupKey][]FileRow{}

	if len(stmt.GroupBy) == 0 {
		groups[""] = rows
		order = []groupKey{""}
	} else {
		for _, row := range rows {
			key := groupKeyFor(row, stmt.GroupBy)
			if _, exists := groups[key]; !exists {
				order = append(order, key)
			}
			groups[key] = append(groups[key], row)
		}
	}

	var result []ResultRow
	for _, key := range order {
		r, err := computeAggregates(groups[key], stmt.Columns)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func groupKeyFor(row FileRow, groupBy []string) string {
	parts := make([]string, len(groupBy))
	for i, col := range groupBy {
		parts[i] = fmt.Sprintf("%v", row.Get(col))
	}
	return strings.Join(parts, "\x00")
}

func computeAggregates(rows []FileRow, cols []parser.ColExpr) (ResultRow, error) {
	r := make(ResultRow, len(cols))
	for _, c := range cols {
		if c.Agg == "" {
			if len(rows) == 0 {
				r[c.String()] = nil
				continue
			}
			val := rows[0].Get(c.Col)
			if val == nil {
				return nil, fmt.Errorf("unknown column %q", c.Col)
			}
			r[c.String()] = val
		} else {
			val, err := computeAgg(c.Agg, c.Col, rows)
			if err != nil {
				return nil, err
			}
			r[c.String()] = val
		}
	}
	return r, nil
}

func computeAgg(agg, col string, rows []FileRow) (interface{}, error) {
	switch agg {
	case "COUNT":
		return int64(len(rows)), nil
	case "SUM":
		var sum int64
		for _, row := range rows {
			v, err := toInt64(row.Get(col))
			if err != nil {
				return nil, fmt.Errorf("SUM(%s): %w", col, err)
			}
			sum += v
		}
		return sum, nil
	case "MIN":
		if len(rows) == 0 {
			return nil, nil
		}
		min := rows[0].Get(col)
		for _, row := range rows[1:] {
			v := row.Get(col)
			if lessVal(v, min) {
				min = v
			}
		}
		return min, nil
	case "MAX":
		if len(rows) == 0 {
			return nil, nil
		}
		max := rows[0].Get(col)
		for _, row := range rows[1:] {
			v := row.Get(col)
			if lessVal(max, v) {
				max = v
			}
		}
		return max, nil
	}
	return nil, fmt.Errorf("unknown aggregate function %q", agg)
}

// lessVal reports whether a < b for comparable types (int64, string, time.Time).
func lessVal(a, b interface{}) bool {
	switch av := a.(type) {
	case int64:
		if bv, ok := b.(int64); ok {
			return av < bv
		}
	case string:
		if bv, ok := b.(string); ok {
			return av < bv
		}
	case time.Time:
		if bv, ok := b.(time.Time); ok {
			return av.Before(bv)
		}
	}
	return false
}
