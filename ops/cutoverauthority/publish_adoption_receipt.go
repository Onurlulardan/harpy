package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func publishAdoptionReceipt(
	profile adoptionInvocation, signed signedAdoptionReceipt,
) error {
	if err := publishNoReplace(profile.ReceiptPath, signed.document); err != nil {
		return fmt.Errorf("authority receipt publication failed")
	}
	if err := publishNoReplace(
		profile.ReceiptSignaturePath, signed.signatureDocument,
	); err != nil {
		return fmt.Errorf("authority receipt signature publication failed")
	}
	return nil
}

func publishNoReplace(path string, document []byte) error {
	if len(document) == 0 {
		return fmt.Errorf("artifact document is empty")
	}
	directory := filepath.Dir(path)
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("artifact output directory is unavailable")
	}
	if _, err := os.Lstat(path); err == nil {
		if !existingArtifactMatches(path, document) {
			return fmt.Errorf("existing artifact differs")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("artifact target inspection failed")
	}
	temporary, err := os.CreateTemp(directory, ".cutover-authority-*")
	if err != nil {
		return fmt.Errorf("artifact staging failed")
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := stageArtifact(temporary, document); err != nil {
		return fmt.Errorf("artifact staging failed")
	}
	if err := os.Link(temporaryPath, path); err != nil {
		if existingArtifactMatches(path, document) {
			return nil
		}
		return fmt.Errorf("artifact no-replace publication failed")
	}
	if runtime.GOOS != "windows" {
		directoryHandle, openErr := os.Open(directory)
		if openErr != nil {
			return fmt.Errorf("artifact directory sync failed")
		}
		syncErr := directoryHandle.Sync()
		closeErr := directoryHandle.Close()
		if syncErr != nil || closeErr != nil {
			return fmt.Errorf("artifact directory sync failed")
		}
	}
	return nil
}

type artifactStagingFile interface {
	Chmod(os.FileMode) error
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

func stageArtifact(file artifactStagingFile, document []byte) error {
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return err
	}
	written, err := file.Write(document)
	if err != nil || written != len(document) {
		return fmt.Errorf("artifact write was incomplete")
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closed = true
	return nil
}

func existingArtifactMatches(path string, document []byte) bool {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false
	}
	current, err := os.ReadFile(path)
	return err == nil && bytes.Equal(current, document)
}
