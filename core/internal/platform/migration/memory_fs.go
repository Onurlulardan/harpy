package migration

import (
	"io/fs"
	"path"
	"sort"
)

type memoryFS struct {
	files       map[string][]byte
	directories map[string][]memoryEntry
}

func newMemoryFS(documents map[string][]byte) *memoryFS {
	result := &memoryFS{
		files:       make(map[string][]byte, len(documents)),
		directories: map[string][]memoryEntry{".": nil},
	}
	directoryNames := map[string]struct{}{".": {}}
	for name, content := range documents {
		result.files[name] = append([]byte(nil), content...)
		for directory := path.Dir(name); directory != "."; directory = path.Dir(directory) {
			directoryNames[directory] = struct{}{}
		}
	}
	for directory := range directoryNames {
		result.directories[directory] = nil
	}
	for directory := range directoryNames {
		if directory != "." {
			parent := path.Dir(directory)
			result.directories[parent] = append(
				result.directories[parent], newMemoryEntry(path.Base(directory), 0, true),
			)
		}
	}
	for name, content := range result.files {
		parent := path.Dir(name)
		result.directories[parent] = append(
			result.directories[parent], newMemoryEntry(path.Base(name), int64(len(content)), false),
		)
	}
	for directory := range result.directories {
		sort.Slice(result.directories[directory], func(left, right int) bool {
			return result.directories[directory][left].name < result.directories[directory][right].name
		})
	}
	return result
}

func (source *memoryFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, pathError("open", name, fs.ErrInvalid)
	}
	if entries, exists := source.directories[name]; exists {
		return newMemoryDirectory(name, entries), nil
	}
	content, exists := source.files[name]
	if !exists {
		return nil, pathError("open", name, fs.ErrNotExist)
	}
	return newMemoryFile(name, content), nil
}

func (source *memoryFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, exists := source.directories[name]
	if !exists {
		return nil, pathError("readdir", name, fs.ErrNotExist)
	}
	return cloneDirectoryEntries(entries), nil
}

func pathError(operation, name string, cause error) error {
	return &fs.PathError{Op: operation, Path: name, Err: cause}
}
