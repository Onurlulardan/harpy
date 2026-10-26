package platformops

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

func readRecoveryRuntimeFile(
	path string,
	maximum int64,
	sensitive bool,
) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > maximum ||
		sensitive && runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("recovery artifact is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recovery artifact is unavailable")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("recovery artifact changed while opening")
	}
	document, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(document)) > maximum {
		clear(document)
		return nil, fmt.Errorf("recovery artifact exceeds its limit")
	}
	return document, nil
}
