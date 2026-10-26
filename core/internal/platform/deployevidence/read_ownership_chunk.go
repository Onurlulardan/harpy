package deployevidence

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ownershipChunkReader struct {
	baseDirectory string
	relativePath  string
	maximumBytes  int64
	file          *os.File
	completed     bool
}

func (reader *ownershipChunkReader) Read(target []byte) (int, error) {
	if reader.completed {
		return 0, io.EOF
	}
	if reader.file == nil {
		file, err := reader.open()
		if err != nil {
			return 0, err
		}
		reader.file = file
	}
	count, err := reader.file.Read(target)
	if err != nil {
		_ = reader.file.Close()
		reader.file = nil
		reader.completed = true
	}
	return count, err
}

func (reader *ownershipChunkReader) open() (*os.File, error) {
	if reader.baseDirectory == "" || reader.relativePath == "" || reader.maximumBytes < 1 {
		return nil, fmt.Errorf("ownership chunk source is invalid")
	}
	candidate := filepath.Join(reader.baseDirectory, filepath.FromSlash(reader.relativePath))
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil || !pathWithin(reader.baseDirectory, resolved) {
		return nil, fmt.Errorf("ownership chunk is unavailable")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("ownership chunk is unavailable")
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > reader.maximumBytes {
		_ = file.Close()
		return nil, fmt.Errorf("ownership chunk is not a bounded regular file")
	}
	return file, nil
}

func pathWithin(baseDirectory, candidate string) bool {
	relative, err := filepath.Rel(baseDirectory, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(
		relative, ".."+string(filepath.Separator),
	)
}

var _ io.Reader = (*ownershipChunkReader)(nil)
