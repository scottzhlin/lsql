package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/scottlin/lsql/internal/evaluator"
	"github.com/scottlin/lsql/internal/formatter"
	"github.com/scottlin/lsql/internal/lexer"
	"github.com/scottlin/lsql/internal/parser"
)

func main() {
	format := flag.String("format", "table", "output format: table, csv, json")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		query := strings.Join(args, " ")
		if err := runQuery(query, formatter.Format(*format)); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	repl(formatter.Format(*format))
}

func repl(format formatter.Format) {
	fmt.Println("lsql — enter SQL queries (or 'exit' to quit)")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		if err := runQuery(line, format); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
}

func runQuery(query string, format formatter.Format) error {
	tokens, err := lexer.New(query).Tokenize()
	if err != nil {
		return err
	}
	stmt, err := parser.New(tokens).Parse()
	if err != nil {
		return err
	}
	result, err := evaluator.Run(stmt)
	if err != nil {
		return err
	}
	return formatter.Print(os.Stdout, result, format)
}
