package evaluator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Scan reads directory entries at path. If recursive is true, it descends into subdirectories.
// Pass nil for stats to print warnings immediately (legacy behavior for direct callers).
func Scan(path string, recursive bool, stats *scanStats) ([]FileRow, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", path)
	}
	if recursive {
		return scanRecursive(path, stats)
	}
	return scanFlat(path, stats)
}

func scanFlat(path string, stats *scanStats) ([]FileRow, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	rows := make([]FileRow, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			stats.warnf("skipping %q: %v", filepath.Join(path, e.Name()), err)
			continue
		}
		rows = append(rows, newFileRow(filepath.Join(path, e.Name()), info, 0))
	}
	return rows, nil
}

func scanRecursive(root string, stats *scanStats) ([]FileRow, error) {
	var rows []FileRow
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			stats.warnf("skipping %q: %v", path, err)
			return nil
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		depth := strings.Count(rel, string(filepath.Separator))
		info, err := d.Info()
		if err != nil {
			stats.warnf("skipping %q: %v", path, err)
			return nil
		}
		rows = append(rows, newFileRow(path, info, depth))
		return nil
	})
	return rows, err
}
