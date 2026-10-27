package attestation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func PreflightWriteNewDirectory(directory string) (returnErr error) {
	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory {
		return fmt.Errorf("attestation directory must be absolute and clean")
	}
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("attestation directory is unavailable")
	}
	temporary, err := os.CreateTemp(directory, ".attestation-preflight-*")
	if err != nil {
		return fmt.Errorf("create attestation preflight file: %w", err)
	}
	temporaryPath := temporary.Name()
	linkedPath := temporaryPath + ".publish"
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		for _, path := range []string{linkedPath, temporaryPath} {
			if removeErr := os.Remove(path); returnErr == nil &&
				removeErr != nil && !os.IsNotExist(removeErr) {
				returnErr = fmt.Errorf("remove attestation preflight file: %w", removeErr)
			}
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("secure attestation preflight file: %w", err)
	}
	if _, err := temporary.Write([]byte("preflight")); err != nil {
		return fmt.Errorf("write attestation preflight file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync attestation preflight file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close attestation preflight file: %w", err)
	}
	closed = true
	if err := os.Link(temporaryPath, linkedPath); err != nil {
		return fmt.Errorf("link attestation preflight file: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}
