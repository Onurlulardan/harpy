package storagecatalog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func openRepositoryRoot(root string) (*os.Root, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, fmt.Errorf("storage catalog repository root must be absolute and clean")
	}
	expected, err := inspectAbsoluteDirectory(root)
	if err != nil {
		return nil, err
	}
	repository, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open storage catalog repository root: %w", err)
	}
	opened, err := repository.Stat(".")
	current, currentErr := inspectAbsoluteDirectory(root)
	if err != nil || currentErr != nil || !os.SameFile(expected, opened) ||
		!os.SameFile(expected, current) {
		repository.Close()
		return nil, fmt.Errorf("storage catalog repository root changed while opening")
	}
	return repository, nil
}

func inspectAbsoluteDirectory(root string) (os.FileInfo, error) {
	volume := filepath.VolumeName(root)
	remainder := strings.TrimPrefix(root, volume)
	separator := string(filepath.Separator)
	current := volume + separator
	remainder = strings.TrimLeft(remainder, separator)
	parts := []string{}
	if remainder != "" {
		parts = strings.Split(remainder, separator)
	}
	if len(parts) == 0 {
		parts = []string{""}
	}
	var info os.FileInfo
	for _, part := range parts {
		if part != "" {
			current = filepath.Join(current, part)
		}
		candidate, err := os.Lstat(current)
		if err != nil {
			return nil, fmt.Errorf("inspect storage catalog repository root: %w", err)
		}
		if candidate.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("storage catalog repository root contains a symbolic link")
		}
		if !candidate.IsDir() {
			return nil, fmt.Errorf("storage catalog repository root must be a directory")
		}
		info = candidate
	}
	return info, nil
}

func readBoundedFile(repository *os.Root, relative string, maximum int64) ([]byte, error) {
	info, err := inspectRepositoryPath(repository, relative)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maximum {
		return nil, fmt.Errorf("%s must be a bounded regular file", relative)
	}
	file, err := openInspectedFile(repository, relative, info)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	if int64(len(content)) > maximum {
		return nil, fmt.Errorf("%s exceeds its byte limit", relative)
	}
	if err := verifyOpenFileStable(repository, relative, file, info); err != nil {
		return nil, err
	}
	return content, nil
}

func inspectRepositoryPath(repository *os.Root, relative string) (os.FileInfo, error) {
	parts := strings.Split(filepath.FromSlash(relative), string(filepath.Separator))
	current := ""
	var info os.FileInfo
	for index, part := range parts {
		current = filepath.Join(current, part)
		candidate, err := repository.Lstat(current)
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", relative, err)
		}
		if candidate.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("%s contains a symbolic link", relative)
		}
		if index < len(parts)-1 && !candidate.IsDir() {
			return nil, fmt.Errorf("%s has a non-directory parent", relative)
		}
		info = candidate
	}
	return info, nil
}

func openInspectedFile(
	repository *os.Root, relative string, expected os.FileInfo,
) (*os.File, error) {
	file, err := repository.Open(filepath.FromSlash(relative))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", relative, err)
	}
	opened, err := file.Stat()
	if err != nil || !os.SameFile(expected, opened) {
		file.Close()
		return nil, fmt.Errorf("%s changed while opening", relative)
	}
	return file, nil
}

func verifyOpenFileStable(
	repository *os.Root, relative string, file *os.File, expected os.FileInfo,
) error {
	opened, openErr := file.Stat()
	current, pathErr := repository.Lstat(filepath.FromSlash(relative))
	if openErr != nil || pathErr != nil || !os.SameFile(expected, opened) ||
		!os.SameFile(expected, current) || opened.Size() != expected.Size() ||
		!opened.ModTime().Equal(expected.ModTime()) {
		return fmt.Errorf("%s changed while reading", relative)
	}
	return nil
}
