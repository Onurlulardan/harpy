package migration

import (
	"bytes"
	"io/fs"
	"path"
)

type memoryFile struct {
	name   string
	reader *bytes.Reader
	closed bool
}

func newMemoryFile(name string, content []byte) *memoryFile {
	return &memoryFile{name: name, reader: bytes.NewReader(content)}
}

func (file *memoryFile) Stat() (fs.FileInfo, error) {
	if file.closed {
		return nil, pathError("stat", file.name, fs.ErrClosed)
	}
	return newMemoryEntry(path.Base(file.name), file.reader.Size(), false), nil
}

func (file *memoryFile) Read(target []byte) (int, error) {
	if file.closed {
		return 0, pathError("read", file.name, fs.ErrClosed)
	}
	return file.reader.Read(target)
}

func (file *memoryFile) Close() error {
	file.closed = true
	return nil
}
