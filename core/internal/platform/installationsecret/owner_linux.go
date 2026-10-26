//go:build linux

package installationsecret

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func createOwnerOnly(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func openOwnerOnly(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	var stat unix.Stat_t
	if err = unix.Fstat(fd, &stat); err != nil || stat.Mode&0o777 != 0o600 ||
		stat.Uid != uint32(os.Geteuid()) || stat.Mode&syscall.S_IFMT != syscall.S_IFREG {
		_ = file.Close()
		return nil, ErrFenced
	}
	return file, nil
}

func publishNoReplace(staging, target string) error {
	err := unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, target, unix.RENAME_NOREPLACE)
	if err == nil {
		return nil
	}
	if err == unix.EEXIST {
		return ErrExists
	}
	return ErrInterrupted
}
