//go:build goexperiment.jsonv2

package jsonutil

import (
	"io"

	json "encoding/json/v2"
	"encoding/json/jsontext"
)

// EncodeIndented writes v as pretty-printed JSON to w, followed by a newline.
func EncodeIndented(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	val := jsontext.Value(b)
	if err := val.Indent(); err != nil {
		return err
	}
	if _, err := w.Write(val); err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	return err
}
