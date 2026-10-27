package installationsecret

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

func validatedTarget(path string, mustExist bool) (string, error) {
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", ErrInvalid
	}
	parent := filepath.Dir(path)
	resolved, err := filepath.EvalSymlinks(parent)
	if err != nil || !samePath(resolved, parent) {
		return "", ErrFenced
	}
	info, err := os.Stat(parent)
	if err != nil || !info.IsDir() {
		return "", ErrFenced
	}
	_, err = os.Lstat(path)
	if mustExist && err != nil {
		return "", ErrFenced
	}
	if !mustExist && err == nil {
		return "", ErrExists
	}
	if err != nil && !os.IsNotExist(err) {
		return "", ErrFenced
	}
	return path, nil
}

func stagingPath(target string) (string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", ErrInterrupted
	}
	name := "." + filepath.Base(target) + ".staging-" + hex.EncodeToString(nonce[:])
	clear(nonce[:])
	return filepath.Join(filepath.Dir(target), name), nil
}

func samePath(left, right string) bool {
	left, leftErr := filepath.Abs(left)
	right, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}
