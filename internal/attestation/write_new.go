package attestation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func WriteNewCanonical(path string, document []byte) (returnErr error) {
	return WriteNewCanonicalBounded(path, document, maximumDocumentBytes)
}

func WriteNewCanonicalBounded(
	path string,
	document []byte,
	maximumBytes int,
) (returnErr error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return fmt.Errorf("attestation target path must be absolute and clean")
	}
	canonical, err := canonicalizeBounded(document, maximumBytes)
	if err != nil {
		return err
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("create attestation temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			if closeErr := temporary.Close(); returnErr == nil && closeErr != nil {
				returnErr = fmt.Errorf("close attestation temporary file: %w", closeErr)
			}
		}
		if removeErr := os.Remove(temporaryPath); returnErr == nil &&
			removeErr != nil && !os.IsNotExist(removeErr) {
			returnErr = fmt.Errorf("remove attestation temporary file: %w", removeErr)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("secure attestation temporary file: %w", err)
	}
	if _, err := temporary.Write(canonical); err != nil {
		return fmt.Errorf("write attestation temporary file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync attestation temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close attestation temporary file: %w", err)
	}
	closed = true
	if err := os.Link(temporaryPath, path); err != nil {
		return fmt.Errorf("publish attestation without replacement: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open attestation directory: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync attestation directory: %w", err)
	}
	return nil
}
