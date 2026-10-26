package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maximumHandoffBytes = 2 * 1024 * 1024

func readRegular(path string, maximum int64) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("artifact path is not absolute")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 || info.Size() < 1 || info.Size() > maximum {
		return nil, fmt.Errorf("artifact is not a bounded regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(content)) != info.Size() {
		return nil, fmt.Errorf("artifact changed while reading")
	}
	return content, nil
}

func writeAtomic(path string, content []byte) error {
	if !filepath.IsAbs(path) || len(content) == 0 {
		return fmt.Errorf("atomic artifact target is invalid")
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact target directory is invalid")
	}
	pending, err := os.CreateTemp(directory, "."+filepath.Base(path)+".pending-*")
	if err != nil {
		return err
	}
	pendingName := pending.Name()
	defer os.Remove(pendingName)
	if err := pending.Chmod(0o600); err != nil {
		pending.Close()
		return err
	}
	if _, err := pending.Write(content); err != nil {
		pending.Close()
		return err
	}
	if err := pending.Sync(); err != nil {
		pending.Close()
		return err
	}
	if err := pending.Close(); err != nil {
		return err
	}
	return replaceAtomic(pendingName, path)
}
