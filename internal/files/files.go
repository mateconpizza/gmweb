// Package files contains functions for working with files.
package files

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	ErrFileNotFound = errors.New("file not found")
	ErrPathNotFound = errors.New("path not found")
	ErrFileExists   = errors.New("file already exists")
	ErrPathEmpty    = errors.New("path is empty")
)

const (
	dirPerm  = 0o755 // Permissions for new directories.
	FilePerm = 0o644 // Permissions for new files.
)

// Exists checks if a file exists.
func Exists(s string) bool {
	_, err := os.Stat(s)
	return !os.IsNotExist(err)
}

// ExistsErr checks if a file exists and returns an error if it does not.
func ExistsErr(p string) error {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}

		return fmt.Errorf("%w", err)
	}

	return nil
}

// List returns all files found in a given path.
func List(root, pattern string) ([]string, error) {
	query := filepath.Join(root, pattern)

	files, err := filepath.Glob(query)
	if err != nil {
		return nil, fmt.Errorf("%w: getting files query: %q", err, query)
	}

	slog.Debug("found files", "count", len(files), "path", root)

	return files, nil
}

// MkdirAll creates all the given paths.
func MkdirAll(s ...string) error {
	for _, p := range s {
		if p == "" {
			return ErrPathEmpty
		}

		if err := mkdir(p); err != nil {
			return err
		}
	}

	return nil
}

// mkdir creates a new directory at the specified path.
func mkdir(s string) error {
	if Exists(s) {
		return nil
	}

	slog.Debug("creating path", "path", s)

	if err := os.MkdirAll(s, dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", s, err)
	}

	return nil
}

// Remove removes the specified file if it exists.
func Remove(s string) error {
	if !Exists(s) {
		return fmt.Errorf("%w: %q", ErrFileNotFound, s)
	}

	slog.Debug("removing path", "path", s)

	if err := os.Remove(s); err != nil {
		return fmt.Errorf("removing file: %w", err)
	}

	return nil
}

// RemoveAll removes the specified file if it exists.
func RemoveAll(s string) error {
	if !Exists(s) {
		return fmt.Errorf("%w: %q", ErrFileNotFound, s)
	}

	slog.Debug("removing path", "path", s)

	if err := os.RemoveAll(s); err != nil {
		return fmt.Errorf("removing file: %w", err)
	}

	return nil
}

func Rename(oldPath, newName string) error {
	if !Exists(oldPath) {
		return fmt.Errorf("%w: %q", ErrFileNotFound, oldPath)
	}

	basePath := filepath.Dir(oldPath)
	newPath := filepath.Join(basePath, newName)

	return os.Rename(oldPath, newPath)
}

// EnsureSuffix appends the specified suffix to the filename.
func EnsureSuffix(s, suffix string) string {
	if s == "" {
		return s
	}

	e := filepath.Ext(s)
	if e == suffix || e != "" {
		return s
	}

	return fmt.Sprintf("%s%s", s, suffix)
}

// StripSuffixes removes all suffixes from the path.
func StripSuffixes(p string) string {
	for ext := filepath.Ext(p); ext != ""; ext = filepath.Ext(p) {
		p = p[:len(p)-len(ext)]
	}

	return p
}

// Touch creates a file at this given path.
// If the file already exists, the function succeeds when exist_ok is true.
func Touch(s string, existsOK bool) (*os.File, error) {
	if Exists(s) && !existsOK {
		return nil, fmt.Errorf("%w: %q", ErrFileExists, s)
	}

	if !Exists(filepath.Dir(s)) {
		if err := MkdirAll(filepath.Dir(s)); err != nil {
			return nil, err
		}
	}

	f, err := os.Create(s)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return f, nil
}

// FindByExtList returns a list of files with the specified extension in the
// given directory.
func FindByExtList(root string, ext ...string) ([]string, error) {
	if !Exists(root) {
		slog.Warn("path not found", "path", root)
		return nil, ErrPathNotFound
	}

	var files []string

	for _, e := range ext {
		f, err := findByExt(root, e)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		files = append(files, f...)
	}

	return files, nil
}

func findByExt(root, ext string) ([]string, error) {
	if !Exists(root) {
		slog.Error("path not found", "path", root)
		return nil, ErrPathNotFound
	}

	files, err := filepath.Glob(root + "/*" + ext)
	if err != nil {
		return nil, fmt.Errorf("getting files: %w with suffix: %q", err, ext)
	}

	return files, nil
}

// PrioritizeFile moves a file to the front of the list.
func PrioritizeFile(files []string, name string) {
	if len(files) == 0 {
		return
	}

	for i, f := range files {
		if filepath.Base(f) == name {
			if i != 0 {
				files[0], files[i] = files[i], files[0]
			}
			break
		}
	}
}
