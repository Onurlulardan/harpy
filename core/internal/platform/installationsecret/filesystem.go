package installationsecret

import (
	"os"
	"runtime"
)

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	err = directory.Sync()
	closeErr := directory.Close()
	if runtime.GOOS == "windows" {
		err = nil
	}
	if err != nil {
		return err
	}
	return closeErr
}
