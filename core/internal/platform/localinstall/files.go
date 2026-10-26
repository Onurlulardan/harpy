package localinstall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const MaximumArtifactBytes = 1 << 20

func ReadRegular(path string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > maximum {
		return nil, fmt.Errorf("local installation artifact is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(content)) > maximum {
		clear(content)
		return nil, fmt.Errorf("local installation artifact is invalid")
	}
	return content, nil
}

func WriteExclusive(path string, content []byte, mode os.FileMode) error {
	if len(content) == 0 || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return fmt.Errorf("local installation artifact path is invalid")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err = file.Write(content); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		return err
	}
	return syncParent(path)
}

func RemoveRegular(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("local installation artifact is invalid")
	}
	if err = os.Remove(path); err != nil {
		return err
	}
	return syncParent(path)
}

func syncParent(path string) error {
	directory, err := os.Open(filepath.Dir(path))
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
