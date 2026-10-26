package installationsecret

import (
	"os"
	"path/filepath"
)

func Remove(path string) error {
	target, err := validatedTarget(path, true)
	if err != nil {
		return err
	}
	file, err := openOwnerOnly(target)
	if err != nil {
		return ErrFenced
	}
	opened, statErr := file.Stat()
	current, pathErr := os.Lstat(target)
	closeErr := file.Close()
	if statErr != nil || pathErr != nil || closeErr != nil ||
		!opened.Mode().IsRegular() || !os.SameFile(opened, current) {
		return ErrFenced
	}
	if err = os.Remove(target); err != nil {
		return ErrInterrupted
	}
	if err = syncDirectory(filepath.Dir(target)); err != nil {
		return ErrInterrupted
	}
	return nil
}
