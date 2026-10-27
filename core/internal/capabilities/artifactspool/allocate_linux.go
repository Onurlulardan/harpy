//go:build linux

package artifactspool

import (
	"os"

	"golang.org/x/sys/unix"
)

func allocateFile(file *os.File, size int64) error {
	if err := unix.Fallocate(int(file.Fd()), 0, 0, size); err != nil {
		return err
	}
	return file.Sync()
}
