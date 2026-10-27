//go:build windows

package artifactspool

import "os"

func allocateFile(file *os.File, size int64) error {
	if err := file.Truncate(size); err != nil {
		return err
	}
	return file.Sync()
}
