package devjournal

import (
	"fmt"
	"os"
)

func acquireFileLock(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("development authority journal is locked")
	}
	return file, nil
}

func releaseFileLock(file *os.File, path string) {
	_ = file.Close()
	_ = os.Remove(path)
}
