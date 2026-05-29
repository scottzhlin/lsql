package evaluator

import (
	"fmt"
	"os"
	"time"
)

// Meta holds query execution statistics for CLI feedback.
type Meta struct {
	From      string
	Recursive bool
	Scanned   int // entries read from filesystem
	Matched   int // rows after WHERE (before GROUP BY / LIMIT)
	Duration  time.Duration
	Warnings  int
}

// scanStats collects per-query scan/filter warnings. Nil disables accounting
// and preserves legacy behavior (warnings always printed to stderr).
type scanStats struct {
	verbose bool
	warns   int
}

func (s *scanStats) warnf(format string, args ...interface{}) {
	if s == nil {
		fmt.Fprintf(os.Stderr, "warning: "+format+"\n", args...)
		return
	}
	s.warns++
	if s.verbose {
		fmt.Fprintf(os.Stderr, "warning: "+format+"\n", args...)
	}
}
