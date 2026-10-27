package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

var (
	renamePath = os.Rename
	randomRead = rand.Read
)

func writePrivate(path string, data []byte) error {
	temp, err := createPrivateTemp(filepath.Dir(path))
	if err != nil {
		return err
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tempName)
		}
	}()
	if _, err = temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = replacePrivate(tempName, path); err != nil {
		return err
	}
	committed = true
	return nil
}

func createPrivateTemp(directory string) (*os.File, error) {
	for range 100 {
		name, err := siblingName(directory, ".tmp-")
		if err != nil {
			return nil, err
		}
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		return file, err
	}
	return nil, errors.New("private temp file was not created")
}

func siblingName(directory, prefix string) (string, error) {
	var buf [8]byte
	if _, err := randomRead(buf[:]); err != nil {
		return "", err
	}
	return filepath.Join(directory, prefix+hex.EncodeToString(buf[:])), nil
}

func replacePrivate(tempName, path string) error {
	if runtime.GOOS != "windows" {
		return renamePath(tempName, path)
	}
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return renamePath(tempName, path)
	}
	if err != nil {
		return err
	}
	backup, err := siblingName(filepath.Dir(path), ".bak-")
	if err != nil {
		return err
	}
	if err = renamePath(path, backup); err != nil {
		return err
	}
	if err = renamePath(tempName, path); err != nil {
		if restoreErr := renamePath(backup, path); restoreErr != nil {
			return errors.Join(err, restoreErr)
		}
		return err
	}
	return os.Remove(backup)
}
