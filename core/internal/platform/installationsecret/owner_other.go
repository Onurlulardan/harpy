//go:build !linux && !windows

package installationsecret

import (
	"os"
	"runtime"
)

func createOwnerOnly(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
}

func openOwnerOnly(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() ||
		info.Mode().Perm() != 0o600 {
		return nil, ErrFenced
	}
	return os.Open(path)
}

func publishNoReplace(staging, target string) error {
	if err := os.Link(staging, target); err != nil {
		if _, statErr := os.Lstat(target); statErr == nil {
			return ErrExists
		}
		return ErrInterrupted
	}
	if err := os.Remove(staging); err != nil && runtime.GOOS != "windows" {
		return ErrInterrupted
	}
	return nil
}
