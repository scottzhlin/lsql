package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolvePath(p string) (string, error) {
	switch p {
	case ".":
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot resolve current directory: %w", err)
		}
		return cwd, nil
	case "~":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		return home, nil
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		return filepath.Join(home, p[2:]), nil
	}
	return filepath.Clean(p), nil
}
