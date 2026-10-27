package licenseauthoritysource

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func validArtifactPath(path string) bool {
	return path != "" && filepath.IsAbs(path) && filepath.Clean(path) == path
}

func readStableArtifact(path string, maximum int64, ownerOnly bool) ([]byte, error) {
	if !validArtifactPath(path) || maximum < 1 {
		return nil, ErrInvalidConfiguration
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 ||
		before.Size() < 1 || before.Size() > maximum {
		return nil, fmt.Errorf("license artifact is invalid")
	}
	if ownerOnly && runtime.GOOS != "windows" && before.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("license artifact permissions are not owner-only")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open license artifact: %w", err)
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return nil, fmt.Errorf("license artifact changed during open")
	}
	document, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || len(document) == 0 || int64(len(document)) > maximum {
		return nil, fmt.Errorf("read license artifact")
	}
	final, err := file.Stat()
	if err != nil || !os.SameFile(after, final) || final.Size() != int64(len(document)) ||
		final.ModTime() != after.ModTime() {
		return nil, fmt.Errorf("license artifact changed during read")
	}
	return document, nil
}
