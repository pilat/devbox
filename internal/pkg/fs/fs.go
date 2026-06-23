package fs

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

type FileSystem interface {
	Getwd() (string, error)
	Stat(path string) (FileInfo, error)
	ReadDir(path string) ([]DirEntry, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	RemoveAll(path string) error
}

// FileInfo mirrors os.FileInfo to enable mocking (mockery has issues with stdlib interfaces).
type FileInfo interface {
	Name() string
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	IsDir() bool
	Sys() any
}

// DirEntry mirrors fs.DirEntry to enable mocking.
type DirEntry interface {
	Name() string
	IsDir() bool
	Type() fs.FileMode
	Info() (FileInfo, error)
}

// osDirEntry wraps os.DirEntry to return our FileInfo interface.
type osDirEntry struct {
	entry os.DirEntry
}

func (e osDirEntry) Name() string      { return e.entry.Name() }
func (e osDirEntry) IsDir() bool       { return e.entry.IsDir() }
func (e osDirEntry) Type() fs.FileMode { return e.entry.Type() }
func (e osDirEntry) Info() (FileInfo, error) {
	info, err := e.entry.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get dir entry info: %w", err)
	}
	return info, nil
}

var _ FileSystem = (*OSFileSystem)(nil)

type OSFileSystem struct{}

func New() *OSFileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) Getwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return dir, nil
}

func (f *OSFileSystem) Stat(path string) (FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	return info, nil
}

func (f *OSFileSystem) ReadDir(path string) ([]DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", path, err)
	}
	result := make([]DirEntry, len(entries))
	for i, e := range entries {
		result[i] = osDirEntry{entry: e}
	}
	return result, nil
}

func (f *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

func (f *OSFileSystem) RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	return nil
}
