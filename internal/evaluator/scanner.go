package evaluator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scan reads directory entries at path. If recursive is true, it descends into subdirectories.
func Scan(path string, recursive bool) ([]FileRow, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", path)
	}
	if recursive {
		return scanRecursive(path)
	}
	return scanFlat(path)
}

func scanFlat(path string) ([]FileRow, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	rows := make([]FileRow, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		rows = append(rows, newFileRow(filepath.Join(path, e.Name()), info, 0))
	}
	return rows, nil
}

func scanRecursive(root string) ([]FileRow, error) {
	var rows []FileRow
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", path, err)
			return nil
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		depth := strings.Count(rel, string(filepath.Separator))
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rows = append(rows, newFileRow(path, info, depth))
		return nil
	})
	return rows, err
}
