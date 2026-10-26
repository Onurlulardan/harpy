package migration

import (
	"io"
	"io/fs"
	"path"
)

type memoryDirectory struct {
	name    string
	entries []memoryEntry
	index   int
	closed  bool
}

func newMemoryDirectory(name string, entries []memoryEntry) *memoryDirectory {
	return &memoryDirectory{name: name, entries: append([]memoryEntry(nil), entries...)}
}

func (directory *memoryDirectory) Stat() (fs.FileInfo, error) {
	if directory.closed {
		return nil, pathError("stat", directory.name, fs.ErrClosed)
	}
	return newMemoryEntry(path.Base(directory.name), 0, true), nil
}

func (directory *memoryDirectory) Read([]byte) (int, error) {
	if directory.closed {
		return 0, pathError("read", directory.name, fs.ErrClosed)
	}
	return 0, pathError("read", directory.name, fs.ErrInvalid)
}

func (directory *memoryDirectory) Close() error {
	directory.closed = true
	return nil
}

func (directory *memoryDirectory) ReadDir(count int) ([]fs.DirEntry, error) {
	if directory.closed {
		return nil, pathError("readdir", directory.name, fs.ErrClosed)
	}
	if directory.index >= len(directory.entries) {
		if count > 0 {
			return nil, io.EOF
		}
		return []fs.DirEntry{}, nil
	}
	end := len(directory.entries)
	if count > 0 && directory.index+count < end {
		end = directory.index + count
	}
	entries := cloneDirectoryEntries(directory.entries[directory.index:end])
	directory.index = end
	return entries, nil
}
