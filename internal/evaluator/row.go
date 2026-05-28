package evaluator

import (
	"os"
	"path/filepath"
	"time"
)

// FileRow is one row in the virtual filesystem table.
type FileRow struct {
	Name      string
	Path      string
	Size      int64
	Mode      string
	Modified  time.Time
	IsDir     bool
	Extension string
	Depth     int
}

func newFileRow(path string, info os.FileInfo, depth int) FileRow {
	ext := ""
	if !info.IsDir() {
		ext = filepath.Ext(info.Name())
	}
	return FileRow{
		Name:      info.Name(),
		Path:      path,
		Size:      info.Size(),
		Mode:      info.Mode().String(),
		Modified:  info.ModTime(),
		IsDir:     info.IsDir(),
		Extension: ext,
		Depth:     depth,
	}
}

// Get returns the value of a named column; returns nil for unknown columns.
func (r FileRow) Get(col string) interface{} {
	switch col {
	case "name":
		return r.Name
	case "path":
		return r.Path
	case "size":
		return r.Size
	case "mode":
		return r.Mode
	case "modified":
		return r.Modified
	case "is_dir":
		return r.IsDir
	case "extension":
		return r.Extension
	case "depth":
		return int64(r.Depth)
	}
	return nil
}
