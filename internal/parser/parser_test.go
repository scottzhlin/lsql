package parser_test

import (
	"testing"

	"github.com/scottlin/lsql/internal/lexer"
	"github.com/scottlin/lsql/internal/parser"
)

func parse(t *testing.T, sql string) *parser.SelectStmt {
	t.Helper()
	tokens, err := lexer.New(sql).Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	stmt, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return stmt
}

func TestParser_SelectStar(t *testing.T) {
	stmt := parse(t, "SELECT * FROM .")
	if len(stmt.Columns) != 1 || stmt.Columns[0].Col != "*" {
		t.Errorf("expected SELECT *, got %v", stmt.Columns)
	}
	if stmt.From != "." {
		t.Errorf("expected FROM ., got %q", stmt.From)
	}
	if stmt.Recursive {
		t.Error("expected non-recursive")
	}
	if stmt.Limit != -1 {
		t.Errorf("expected limit -1, got %d", stmt.Limit)
	}
}

func TestParser_Recursive(t *testing.T) {
	stmt := parse(t, "SELECT name FROM /tmp RECURSIVE")
	if !stmt.Recursive {
		t.Error("expected recursive=true")
	}
	if stmt.From != "/tmp" {
		t.Errorf("expected FROM /tmp, got %q", stmt.From)
	}
}

func TestParser_WhereAndOrderLimit(t *testing.T) {
	stmt := parse(t, "SELECT name, size FROM . WHERE size > 1000 ORDER BY size DESC LIMIT 10")
	if stmt.Where == nil {
		t.Fatal("expected WHERE clause")
	}
	be, ok := stmt.Where.(parser.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Where)
	}
	if be.Col != "size" || be.Op != ">" {
		t.Errorf("unexpected binary expr: col=%q op=%q", be.Col, be.Op)
	}
	if len(stmt.OrderBy) != 1 || stmt.OrderBy[0].Col != "size" || !stmt.OrderBy[0].Desc {
		t.Errorf("unexpected ORDER BY: %v", stmt.OrderBy)
	}
	if stmt.Limit != 10 {
		t.Errorf("expected LIMIT 10, got %d", stmt.Limit)
	}
}

func TestParser_Aggregates(t *testing.T) {
	stmt := parse(t, "SELECT extension, COUNT(*), SUM(size) FROM . RECURSIVE GROUP BY extension ORDER BY SUM(size) DESC")
	if len(stmt.Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(stmt.Columns))
	}
	if stmt.Columns[1].Agg != "COUNT" || stmt.Columns[1].Col != "*" {
		t.Errorf("column[1] = %+v, want COUNT(*)", stmt.Columns[1])
	}
	if stmt.Columns[2].Agg != "SUM" || stmt.Columns[2].Col != "size" {
		t.Errorf("column[2] = %+v, want SUM(size)", stmt.Columns[2])
	}
	if len(stmt.GroupBy) != 1 || stmt.GroupBy[0] != "extension" {
		t.Errorf("unexpected GROUP BY: %v", stmt.GroupBy)
	}
	if len(stmt.OrderBy) != 1 || stmt.OrderBy[0].Col != "SUM(size)" || !stmt.OrderBy[0].Desc {
		t.Errorf("unexpected ORDER BY: %v", stmt.OrderBy)
	}
}

func TestParser_LogicalExpr(t *testing.T) {
	stmt := parse(t, "SELECT name FROM . WHERE size > 100 AND is_dir = true OR name LIKE '%.go'")
	if stmt.Where == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParser_NotExpr(t *testing.T) {
	stmt := parse(t, "SELECT name FROM . WHERE NOT is_dir = true")
	_, ok := stmt.Where.(parser.NotExpr)
	if !ok {
		t.Fatalf("expected NotExpr, got %T", stmt.Where)
	}
}

func TestParser_BetweenExpr(t *testing.T) {
	stmt := parse(t, "SELECT name FROM . WHERE size BETWEEN 100 AND 1000")
	_, ok := stmt.Where.(parser.BetweenExpr)
	if !ok {
		t.Fatalf("expected BetweenExpr, got %T", stmt.Where)
	}
}

func TestParser_Error_MissingFrom(t *testing.T) {
	tokens, _ := lexer.New("SELECT name WHERE size > 0").Tokenize()
	_, err := parser.New(tokens).Parse()
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
