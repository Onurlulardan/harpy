package migration

import (
	"io/fs"
	"time"
)

type memoryEntry struct {
	name      string
	size      int64
	directory bool
}

func newMemoryEntry(name string, size int64, directory bool) memoryEntry {
	return memoryEntry{name: name, size: size, directory: directory}
}

func (entry memoryEntry) Name() string {
	return entry.name
}

func (entry memoryEntry) Size() int64 {
	return entry.size
}

func (entry memoryEntry) Mode() fs.FileMode {
	if entry.directory {
		return fs.ModeDir | 0o555
	}
	return 0o444
}

func (entry memoryEntry) ModTime() time.Time {
	return time.Time{}
}

func (entry memoryEntry) IsDir() bool {
	return entry.directory
}

func (entry memoryEntry) Sys() any {
	return nil
}

func (entry memoryEntry) Type() fs.FileMode {
	return entry.Mode().Type()
}

func (entry memoryEntry) Info() (fs.FileInfo, error) {
	return entry, nil
}

func cloneDirectoryEntries(entries []memoryEntry) []fs.DirEntry {
	cloned := make([]fs.DirEntry, len(entries))
	for index := range entries {
		cloned[index] = entries[index]
	}
	return cloned
}
