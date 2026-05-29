package jsonutil_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/scottzhlin/lsql/internal/jsonutil"
)

func TestEncodeIndented(t *testing.T) {
	var buf bytes.Buffer
	in := []map[string]any{{"name": "foo.go", "size": int64(1024)}}
	if err := jsonutil.EncodeIndented(&buf, in); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"foo.go"`) {
		t.Errorf("expected encoded name, got %q", out)
	}
	if !strings.Contains(out, "\n") {
		t.Error("expected trailing newline")
	}
}
