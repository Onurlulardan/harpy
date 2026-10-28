package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
)

func readSecretFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 24 || info.Size() > 1024 ||
		runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("secret source is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("secret source is unavailable")
	}
	defer file.Close()
	value, err := io.ReadAll(io.LimitReader(file, 1025))
	if err != nil {
		return nil, fmt.Errorf("secret source cannot be read")
	}
	value = bytes.TrimSuffix(value, []byte("\r\n"))
	value = bytes.TrimSuffix(value, []byte("\n"))
	if len(value) < 24 || len(value) > 256 || bytes.IndexByte(value, 0) >= 0 {
		clear(value)
		return nil, fmt.Errorf("secret source value is invalid")
	}
	return value, nil
}
