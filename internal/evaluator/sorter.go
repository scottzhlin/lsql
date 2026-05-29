package evaluator

import (
	"sort"

	"github.com/scottzhlin/lsql/internal/parser"
)

// SortAndLimit applies ORDER BY ordering and LIMIT truncation to result rows.
// limit=-1 means no limit.
func SortAndLimit(rows []ResultRow, orderBy []parser.OrderItem, limit int) []ResultRow {
	if len(orderBy) > 0 {
		sort.SliceStable(rows, func(i, j int) bool {
			for _, o := range orderBy {
				vi := rows[i][o.Col]
				vj := rows[j][o.Col]
				if lessVal(vi, vj) {
					return !o.Desc
				}
				if lessVal(vj, vi) {
					return o.Desc
				}
			}
			return false
		})
	}
	if limit >= 0 && limit < len(rows) {
		return rows[:limit]
	}
	return rows
}
