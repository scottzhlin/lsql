# lsql Design Spec

**Date:** 2026-05-29  
**Status:** Approved

## Overview

`lsql` is a SQL query interface for the Unix filesystem, inspired by `gitql`. It lets users query directory listings using familiar SQL syntax, supporting both one-shot CLI mode and an interactive REPL.

## User Interaction

- **CLI mode**: `lsql "SELECT name, size FROM . WHERE size > 1000 ORDER BY size DESC"`
- **REPL mode**: run `lsql` with no arguments, enter queries interactively (errors do not exit)

Output format is controlled by `--format=table` (default), `--format=csv`, or `--format=json`.

## Architecture

```
SQL string
    ↓
[Lexer]      — tokenizes input into Token stream
    ↓
[Parser]     — recursive descent, produces AST (SelectStmt)
    ↓
[Evaluator]
  ├── [Scanner]    — walks filesystem (os.ReadDir / filepath.WalkDir)
  ├── [Filter]     — applies WHERE expression tree per row
  ├── [Aggregator] — GROUP BY, COUNT/SUM/MIN/MAX
  └── [Sorter]     — ORDER BY + LIMIT
    ↓
[Formatter]  — renders rows as table / CSV / JSON
    ↓
stdout / stderr
```

### Directory Structure

```
lsql/
├── cmd/lsql/main.go         # entry point: CLI arg or REPL
├── internal/
│   ├── lexer/               # Lexer, Token types
│   ├── parser/              # recursive descent Parser + AST node types
│   ├── evaluator/           # Scanner, Filter, Aggregator, Sorter
│   └── formatter/           # table/csv/json output
└── go.mod
```

Each package has a single responsibility and can be tested in isolation.

## Data Model

`FROM <path>` maps to a virtual table. Each row is one file or directory entry.

| Column      | Type   | Description                                      |
|-------------|--------|--------------------------------------------------|
| `name`      | string | filename (no path)                               |
| `path`      | string | full path                                        |
| `size`      | int64  | bytes (0 for directories)                        |
| `mode`      | string | permission string, e.g. `-rwxr-xr-x`            |
| `modified`  | time   | last modification time                           |
| `is_dir`    | bool   | true if directory                                |
| `extension` | string | file extension including dot, e.g. `.go`; empty if none |
| `depth`     | int    | depth relative to FROM path; always 0 without RECURSIVE |

## SQL Grammar (EBNF, simplified)

```
Query     = "SELECT" Columns "FROM" Path ["RECURSIVE"]
            ["WHERE" Expr]
            ["GROUP BY" ColList]
            ["ORDER BY" OrderList]
            ["LIMIT" Number]

Columns   = "*" | ColExpr ("," ColExpr)*
ColExpr   = AggFunc "(" ( Col | "*" ) ")" | Col
AggFunc   = "COUNT" | "SUM" | "MIN" | "MAX"
            -- COUNT(*) is valid; SUM/MIN/MAX require a named column

Expr      = Expr ("AND" | "OR") Expr
           | "NOT" Expr
           | Col Op Value
           | Col "LIKE" String
           | Col "BETWEEN" Value "AND" Value
           | "(" Expr ")"

Op        = "=" | "!=" | "<" | "<=" | ">" | ">="
```

- `RECURSIVE` after `FROM <path>` triggers `filepath.WalkDir`; without it only one level is read.
- `LIKE` supports `%` as wildcard.
- `modified` comparisons accept strings in `2006-01-02` or RFC3339 format.

## AST Core Types

```go
type SelectStmt struct {
    Columns   []ColExpr
    From      string      // directory path
    Recursive bool
    Where     Expr        // nil = no filter
    GroupBy   []string
    OrderBy   []OrderItem
    Limit     int         // -1 = unlimited
}

// Expr is an interface; concrete types:
//   BinaryExpr  { Left Col, Op string, Right Value }
//   LogicalExpr { Left Expr, Op string, Right Expr }  // AND / OR
//   NotExpr     { Inner Expr }
//   LikeExpr    { Col string, Pattern string }
//   BetweenExpr { Col string, Low, High Value }
```

## Error Handling

| Layer      | Behavior |
|------------|----------|
| Lexer      | Returns error with column position: `lexer error at col 12: unexpected character '#'` |
| Parser     | Returns context: `parse error: expected FROM, got WHERE` |
| Evaluator  | Path not found / permission denied → stderr, skip row (non-fatal) |
| Type error | e.g. `size LIKE '%go'` → detected at evaluation, reported to stderr |
| REPL       | All errors print to stderr and await next query; never exit on error |

## Testing Strategy

| Layer      | Approach |
|------------|----------|
| Lexer      | Unit tests for token stream correctness; edge cases (quotes, operators, whitespace) |
| Parser     | Valid SQL → correct AST; invalid SQL → expected error messages |
| Evaluator  | Use `t.TempDir()` to construct known file trees; assert query results |
| Formatter  | Given a fixed row set, assert table/csv/json string output |
| Integration| End-to-end: SQL string → final output content |

All tests use the standard `testing` package; no external test dependencies.

## Example Queries

```sql
-- Files larger than 1 MB in current directory
SELECT name, size FROM . WHERE size > 1048576

-- All Go files recursively, sorted by size
SELECT name, path, size FROM . RECURSIVE
  WHERE extension = '.go' ORDER BY size DESC

-- Count and total size by extension
SELECT extension, COUNT(*), SUM(size) FROM . RECURSIVE
  GROUP BY extension ORDER BY SUM(size) DESC

-- 10 most recently modified files
SELECT name, modified FROM . RECURSIVE
  ORDER BY modified DESC LIMIT 10

-- Directories only
SELECT name, mode FROM . WHERE is_dir = true
```
