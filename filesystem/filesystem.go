// Package filesystem defines the workspace-relative filesystem capability.
package filesystem

import (
	"context"
	"io/fs"
)

// FS provides concurrent-safe access to paths beneath a provider-defined
// workspace root. Paths use slash separators and are relative to that root.
//
// Implementations reject absolute paths, parent traversal, NUL bytes,
// backslashes, and non-root dot segments. They must also keep resolved symlink
// targets inside the workspace boundary.
type FS interface {
	ReadFile(context.Context, string) ([]byte, error)
	WriteFile(context.Context, string, []byte, fs.FileMode) error
	ReadDir(context.Context, string) ([]fs.DirEntry, error)
	Stat(context.Context, string) (fs.FileInfo, error)
	MkdirAll(context.Context, string, fs.FileMode) error
	Remove(context.Context, string) error
	Rename(context.Context, string, string) error
}
