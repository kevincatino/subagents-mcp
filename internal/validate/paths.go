package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrEmptyPath     = errors.New("path is required")
	ErrRelativePath  = errors.New("path must be absolute")
	ErrRootPath      = errors.New("root path is not allowed")
	ErrNotExist      = errors.New("path does not exist")
	ErrNotDirectory  = errors.New("path is not a directory")
	ErrSymlinkEscape = errors.New("path escapes via symlink")
)

// Dir ensures the provided path is absolute, exists, is a directory, and not the filesystem root.
func Dir(p string) (string, error) {
	if p == "" {
		return "", ErrEmptyPath
	}
	if !filepath.IsAbs(p) {
		return "", ErrRelativePath
	}

	cleaned := filepath.Clean(p)
	if cleaned == string(filepath.Separator) {
		return "", ErrRootPath
	}

	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	if !filepath.IsAbs(resolved) {
		return "", ErrRelativePath
	}
	if resolved == string(filepath.Separator) {
		return "", ErrRootPath
	}

	info, err := os.Stat(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNotExist
		}
		return "", fmt.Errorf("stat path: %w", err)
	}
	if !info.IsDir() {
		return "", ErrNotDirectory
	}

	return resolved, nil
}
