# lsql

Query your filesystem with SQL. Inspired by [gitql](https://github.com/filhodanuvem/gitql).

```
$ lsql "SELECT name, size FROM . WHERE extension = '.go' ORDER BY size DESC LIMIT 5"
name                    size
evaluator_test.go       3821
aggregator_test.go      1823
parser_test.go          1712
scanner_test.go         1607
aggregator.go           1534
```

## Install

```bash
go install github.com/scottzhlin/lsql/cmd/lsql@latest
```

Or build from source:

```bash
git clone https://github.com/scottzhlin/lsql.git
cd lsql
go build -o lsql ./cmd/lsql/
```

## Usage

**One-shot query:**
```bash
lsql "SELECT name, size FROM /path/to/dir WHERE size > 1000"
```

**Interactive REPL:**
```bash
lsql
lsql> SELECT name, size FROM . WHERE extension = '.go' LIMIT 5
lsql> schema          # list columns
lsql> help
lsql> exit
```

**Output format:**
```bash
lsql --format=csv  "SELECT name, size FROM ."
lsql --format=json "SELECT name, size FROM ."
lsql --format=table "SELECT name, size FROM ."   # default
```

**CLI quality-of-life** (summary on stderr, empty results without misleading headers):
```bash
lsql "SELECT name FROM . WHERE extension = '.xyz'"   # → (empty) · 0 rows · scanned N entries
lsql -q "SELECT name FROM ."                         # quiet: no summary line
lsql --headers=always "SELECT name FROM . LIMIT 0" # force CSV/table headers
lsql --fail-on-empty "SELECT name FROM . WHERE false"  # exit code 2 if no rows
lsql --verbose "SELECT name FROM ~ RECURSIVE LIMIT 5"  # print skip warnings live
```

## Virtual Table

Every directory is a queryable table. Each file or directory entry is one row.

| Column      | Type   | Description                              |
|-------------|--------|------------------------------------------|
| `name`      | string | Filename (without path)                  |
| `path`      | string | Full path                                |
| `size`      | int    | Size in bytes (0 for directories)        |
| `mode`      | string | Permission string, e.g. `-rwxr-xr-x`    |
| `modified`  | time   | Last modification time                   |
| `is_dir`    | bool   | `true` if directory                      |
| `extension` | string | File extension including dot, e.g. `.go` |
| `depth`     | int    | Depth relative to FROM path (RECURSIVE)  |

## SQL Reference

### SELECT

```sql
-- All columns
SELECT * FROM .

-- Specific columns
SELECT name, size, extension FROM /tmp

-- Aggregate functions
SELECT COUNT(*), SUM(size), MIN(size), MAX(size) FROM .
```

### FROM

```sql
-- Current directory (flat, non-recursive)
SELECT name FROM .

-- Absolute path
SELECT name FROM /usr/local/bin

-- Recursive traversal
SELECT name, depth FROM . RECURSIVE
```

### WHERE

```sql
-- Comparison operators: = != < <= > >=
SELECT name FROM . WHERE size > 1048576

-- String comparison
SELECT name FROM . WHERE extension = '.go'

-- LIKE with wildcards (% = any sequence, _ = any single character)
SELECT name FROM . WHERE name LIKE '%.go'
SELECT name FROM . WHERE name LIKE 'main_go'

-- BETWEEN
SELECT name FROM . WHERE size BETWEEN 1000 AND 9999

-- Boolean
SELECT name FROM . WHERE is_dir = true

-- Logical operators: AND OR NOT
SELECT name FROM . WHERE extension = '.go' AND size > 10000
SELECT name FROM . WHERE is_dir = true OR size = 0
SELECT name FROM . WHERE NOT is_dir = true
```

### GROUP BY / ORDER BY / LIMIT

```sql
-- Group and aggregate
SELECT extension, COUNT(*), SUM(size) FROM . RECURSIVE
  GROUP BY extension
  ORDER BY SUM(size) DESC

-- Sort and paginate
SELECT name, size FROM . RECURSIVE
  ORDER BY modified DESC
  LIMIT 10
```

## Examples

```bash
# Find the 10 largest files anywhere under home
lsql "SELECT name, path, size FROM ~ RECURSIVE ORDER BY size DESC LIMIT 10"

# Count files by extension
lsql "SELECT extension, COUNT(*) FROM . RECURSIVE GROUP BY extension ORDER BY COUNT(*) DESC"

# Files modified recently (sort by modification time)
lsql "SELECT name, modified FROM . RECURSIVE ORDER BY modified DESC LIMIT 20"

# Find all Go files larger than 100 KB
lsql "SELECT name, path, size FROM . RECURSIVE WHERE extension = '.go' AND size > 102400"

# Show only directories
lsql "SELECT name, mode FROM . WHERE is_dir = true"

# Export to CSV for spreadsheet analysis
lsql --format=csv "SELECT name, size, extension FROM . RECURSIVE" > files.csv

# JSON output for further processing
lsql --format=json "SELECT name, size FROM . WHERE size > 1000" | jq '.[] | .name'
```

## Architecture

```
SQL string
    ↓ Lexer      (internal/lexer)     — tokenization
    ↓ Parser     (internal/parser)    — recursive descent → AST
    ↓ Evaluator  (internal/evaluator) — scan + filter + aggregate + sort
    ↓ Formatter  (internal/formatter) — table / CSV / JSON
```

Standard library only — no external dependencies.

## License

MIT
