package recoverysigning

import (
	"os"
	"runtime"
)

func prepareStaging(staging string) error {
	info, err := os.Lstat(staging)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return ErrFenced
	}
	if err := os.RemoveAll(staging); err != nil {
		return ErrFenced
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return ErrInterrupted
	}
	err = directory.Sync()
	closeErr := directory.Close()
	if runtime.GOOS == "windows" {
		err = nil
	}
	if err != nil || closeErr != nil {
		return ErrInterrupted
	}
	return nil
}
