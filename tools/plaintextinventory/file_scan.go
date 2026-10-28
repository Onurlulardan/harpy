package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func scanPaths(
	ctx context.Context,
	secret []byte,
	paths []string,
	allowedFile string,
) (int, error) {
	allowed, err := filepath.EvalSymlinks(allowedFile)
	if err != nil {
		return 0, fmt.Errorf("allowed secret source is unavailable")
	}
	allowed, err = filepath.Abs(allowed)
	if err != nil {
		return 0, fmt.Errorf("allowed secret source is invalid")
	}
	verified := 0
	for _, root := range paths {
		err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink is forbidden in scan surface: %s", path)
			}
			if entry.IsDir() {
				return nil
			}
			absolute, absErr := filepath.Abs(path)
			if absErr != nil {
				return absErr
			}
			if filepath.Clean(absolute) == filepath.Clean(allowed) {
				verified++
				return nil
			}
			file, openErr := os.Open(path)
			if openErr != nil {
				return openErr
			}
			found, scanErr := streamContains(ctx, file, secret, 64*1024*1024)
			closeErr := file.Close()
			if scanErr != nil {
				return scanErr
			}
			if closeErr != nil {
				return closeErr
			}
			if found {
				return fmt.Errorf("plaintext secret found in file %s", path)
			}
			verified++
			return nil
		})
		if err != nil {
			return verified, err
		}
	}
	return verified, nil
}
