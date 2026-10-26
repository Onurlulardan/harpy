package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func writeOwnedFile(
	path string,
	content []byte,
	mode fs.FileMode,
	uid int,
	gid int,
	chown ownershipChanger,
) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(path), err)
	}
	if _, err = file.Write(content); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	if err := chown(path, uid, gid); err != nil {
		return fmt.Errorf("own %s: %w", filepath.Base(path), err)
	}
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("protect %s: %w", filepath.Base(path), err)
	}
	return nil
}
