package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func publishExact(directory string, files map[string][]byte) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read recovery target")
	}
	if len(entries) > 0 {
		return verifyExact(directory, entries, files)
	}
	for name, content := range files {
		mode := os.FileMode(0o600)
		if filepath.Ext(name) == ".json" {
			mode = 0o644
		}
		path := filepath.Join(directory, name)
		file, createErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
		if createErr != nil {
			return fmt.Errorf("create recovery projection")
		}
		_, writeErr := file.Write(content)
		if writeErr == nil {
			writeErr = file.Sync()
		}
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			return fmt.Errorf("publish recovery projection")
		}
	}
	return syncDirectory(directory)
}

func verifyExact(
	directory string,
	entries []os.DirEntry,
	files map[string][]byte,
) error {
	if len(entries) != len(files) {
		return fmt.Errorf("recovery projection is fenced")
	}
	for _, entry := range entries {
		expected, exists := files[entry.Name()]
		if !exists || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("recovery projection is fenced")
		}
		actual, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil || !bytes.Equal(actual, expected) {
			return fmt.Errorf("recovery projection is fenced")
		}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open recovery target")
	}
	err = directory.Sync()
	closeErr := directory.Close()
	if runtime.GOOS == "windows" {
		err = nil
	}
	if err != nil || closeErr != nil {
		return fmt.Errorf("sync recovery target")
	}
	return nil
}
