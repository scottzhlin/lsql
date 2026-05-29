//go:build !goexperiment.jsonv2

package jsonutil

import (
	"encoding/json"
	"io"
)

// EncodeIndented writes v as pretty-printed JSON to w, followed by a newline.
func EncodeIndented(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
