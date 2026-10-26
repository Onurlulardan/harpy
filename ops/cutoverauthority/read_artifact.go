package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const (
	maximumSecretBytes   = 64 * 1024
	maximumArtifactBytes = 16 * 1024 * 1024
)

func readArtifact(path string, maximum int) ([]byte, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("artifact is unavailable")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("artifact is unavailable")
	}
	defer file.Close()
	return readOpenedArtifact(file, maximum, false)
}

func readOpenedArtifact(
	file *os.File, maximum int, requireSecretMode bool,
) ([]byte, error) {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 ||
		info.Size() > int64(maximum) {
		return nil, fmt.Errorf("artifact is unavailable or outside its size bound")
	}
	if requireSecretMode && runtime.GOOS != "windows" {
		permissions := info.Mode().Perm()
		if permissions != 0o400 && permissions != 0o600 {
			return nil, fmt.Errorf("secret file permissions are invalid")
		}
	}
	document, err := io.ReadAll(io.LimitReader(file, int64(maximum+1)))
	if err != nil || len(document) < 1 || len(document) > maximum {
		return nil, fmt.Errorf("artifact read failed")
	}
	return document, nil
}

func readSecret(path string) ([]byte, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("secret document is unavailable")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("secret document is unavailable")
	}
	defer file.Close()
	document, err := readOpenedArtifact(file, maximumSecretBytes, true)
	if err != nil {
		return nil, err
	}
	document = bytes.TrimSuffix(document, []byte("\r\n"))
	document = bytes.TrimSuffix(document, []byte("\n"))
	if len(document) == 0 || bytes.IndexByte(document, 0) >= 0 {
		return nil, fmt.Errorf("secret document is invalid")
	}
	return document, nil
}

func bytesReader(document []byte) io.Reader { return bytes.NewReader(document) }
