package parser

// SelectStmt is the root AST node for all lsql queries.
type SelectStmt struct {
	Columns   []ColExpr
	From      string
	Recursive bool
	Where     Expr   // nil = no filter
	GroupBy   []string
	OrderBy   []OrderItem
	Limit     int // -1 = unlimited
}

// ColExpr is one entry in the SELECT list.
type ColExpr struct {
	Agg string // "COUNT", "SUM", "MIN", "MAX", or "" for plain column
	Col string // column name, or "*" for COUNT(*)
}

func (c ColExpr) String() string {
	if c.Agg != "" {
		return c.Agg + "(" + c.Col + ")"
	}
	return c.Col
}

// OrderItem is one entry in ORDER BY.
type OrderItem struct {
	Col  string
	Desc bool
}

// Expr is the interface for all WHERE clause nodes.
type Expr interface {
	exprNode()
}

// BinaryExpr evaluates: col op value  (e.g. size > 1000)
type BinaryExpr struct {
	Col string
	Op  string
	Val interface{} // string, int64, float64, bool, time.Time
}

func (BinaryExpr) exprNode() {}

// LikeExpr evaluates: col LIKE pattern  (% is wildcard)
type LikeExpr struct {
	Col     string
	Pattern string
}

func (LikeExpr) exprNode() {}

// BetweenExpr evaluates: col BETWEEN low AND high
type BetweenExpr struct {
	Col  string
	Low  interface{}
	High interface{}
}

func (BetweenExpr) exprNode() {}

// LogicalExpr combines two expressions with AND or OR.
type LogicalExpr struct {
	Left  Expr
	Op    string // "AND" or "OR"
	Right Expr
}

func (LogicalExpr) exprNode() {}

// NotExpr negates an expression.
type NotExpr struct {
	Inner Expr
}

func (NotExpr) exprNode() {}
