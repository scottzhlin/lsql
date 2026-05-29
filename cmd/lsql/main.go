package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/scottzhlin/lsql/internal/evaluator"
	"github.com/scottzhlin/lsql/internal/formatter"
	"github.com/scottzhlin/lsql/internal/lexer"
	"github.com/scottzhlin/lsql/internal/parser"
)

const version = "0.3.0"

func main() {
	format := flag.String("format", "table", "output format: table, csv, json")
	headers := flag.String("headers", "auto", "column headers: auto, always, never")
	humanSize := flag.Bool("human-size", true, "format size columns as KiB/MiB in output")
	noHumanSize := flag.Bool("no-human-size", false, "print raw byte counts for size columns")
	quiet := flag.Bool("quiet", false, "suppress summary line on stderr")
	quietShort := flag.Bool("q", false, "shorthand for --quiet")
	verbose := flag.Bool("verbose", false, "print filesystem warnings as they occur")
	failOnEmpty := flag.Bool("fail-on-empty", false, "exit with code 2 when a query returns zero rows")
	jsonMeta := flag.Bool("json-meta", false, "with --format=json, wrap output in {\"meta\":...,\"data\":...}")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("lsql", version)
		return
	}
	if *quietShort {
		*quiet = true
	}
	if *noHumanSize {
		*humanSize = false
	}

	headerMode, err := formatter.ParseHeaderMode(*headers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *jsonMeta && formatter.Format(*format) != formatter.JSON {
		fmt.Fprintln(os.Stderr, "error: --json-meta requires --format=json")
		os.Exit(1)
	}

	opts := formatter.Options{
		Format:    formatter.Format(*format),
		Headers:   headerMode,
		HumanSize: *humanSize,
		Quiet:     *quiet,
		JSONMeta:  *jsonMeta,
	}
	runOpts := evaluator.RunOptions{Verbose: *verbose}

	args := flag.Args()
	if len(args) > 0 {
		query := strings.Join(args, " ")
		code, err := runQuery(query, opts, runOpts, *failOnEmpty)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if code == 2 {
			os.Exit(2)
		}
		return
	}

	repl(opts, runOpts)
}

func repl(opts formatter.Options, runOpts evaluator.RunOptions) {
	printBanner()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("lsql> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch strings.ToLower(line) {
		case "exit", "quit", "\\q":
			return
		case "help", "?", "\\h", "\\?":
			printHelp()
			continue
		case "schema", "columns", "\\d":
			printSchema()
			continue
		case "version":
			fmt.Println("lsql", version)
			continue
		}
		if strings.HasPrefix(line, "\\") {
			fmt.Fprintf(os.Stderr, "unknown command %q (type help)\n", line)
			continue
		}
		if _, err := runQuery(line, opts, runOpts, false); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

func printBanner() {
	fmt.Println("lsql", version, "— query the filesystem with SQL")
	fmt.Println("Type help for commands, or exit to quit.")
}

func printHelp() {
	fmt.Print(`Commands:
  help, ?, \h     show this help
  schema, \d      list virtual table columns
  exit, quit, \q  leave the REPL

SQL (subset):
  SELECT cols FROM path [RECURSIVE] [WHERE ...] [GROUP BY ...] [ORDER BY ...] [LIMIT n]

Paths:
  .              current working directory
  ~              home directory
  /path/to/dir   absolute path

Examples:
  SELECT name, size FROM . WHERE extension = '.go' ORDER BY size DESC LIMIT 10
  SELECT extension, COUNT(*) FROM . RECURSIVE GROUP BY extension

CLI flags (also work in one-shot mode):
  --format=table|csv|json   --json-meta (with json)   --headers=auto|always|never
  --human-size / --no-human-size   --quiet (-q)   --verbose   --fail-on-empty
`)
}

func printSchema() {
	cols := []struct{ name, typ, desc string }{
		{"name", "string", "filename (no path)"},
		{"path", "string", "full path"},
		{"size", "int", "size in bytes (0 for directories)"},
		{"mode", "string", "permissions, e.g. -rwxr-xr-x"},
		{"modified", "time", "last modification time"},
		{"is_dir", "bool", "true if directory"},
		{"extension", "string", "file extension including dot, e.g. .go"},
		{"depth", "int", "depth relative to FROM (RECURSIVE only)"},
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "column\ttype\tdescription")
	for _, c := range cols {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", c.name, c.typ, c.desc)
	}
	_ = tw.Flush()
}

// runQuery returns an exit code: 0 ok, 2 empty when failOnEmpty.
func runQuery(query string, opts formatter.Options, runOpts evaluator.RunOptions, failOnEmpty bool) (int, error) {
	tokens, err := lexer.New(query).Tokenize()
	if err != nil {
		return 0, err
	}
	stmt, err := parser.New(tokens).Parse()
	if err != nil {
		return 0, err
	}
	result, err := evaluator.RunWithOptions(stmt, runOpts)
	if err != nil {
		return 0, err
	}
	if err := formatter.Print(os.Stdout, os.Stderr, result, opts); err != nil {
		return 0, err
	}
	if failOnEmpty && len(result.Rows) == 0 {
		return 2, nil
	}
	return 0, nil
}
