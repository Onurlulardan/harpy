package recoveryorchestration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func writeRecoveryArtifactAtomic(
	path string,
	document []byte,
) (err error) {
	target := filepath.Clean(path)
	directory := filepath.Dir(target)
	info, statErr := os.Lstat(directory)
	resolved, resolveErr := filepath.EvalSymlinks(directory)
	if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 ||
		resolveErr != nil || filepath.Clean(resolved) != directory ||
		len(document) == 0 {
		return fmt.Errorf("recovery artifact output path is invalid")
	}
	if existing, existingErr := os.Lstat(target); existingErr == nil {
		if !existing.Mode().IsRegular() ||
			existing.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("recovery artifact output is invalid")
		}
	} else if !os.IsNotExist(existingErr) {
		return fmt.Errorf("recovery artifact output is unavailable")
	}
	file, err := os.CreateTemp(directory, "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create recovery artifact: %w", err)
	}
	temp := file.Name()
	defer func() {
		_ = file.Close()
		if err != nil {
			_ = os.Remove(temp)
		}
	}()
	if err = file.Chmod(0o600); err != nil {
		return err
	}
	if _, err = file.Write(document); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(temp, target); err != nil {
		return fmt.Errorf("publish recovery artifact: %w", err)
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	parent, openErr := os.Open(directory)
	if openErr != nil {
		return openErr
	}
	defer parent.Close()
	return parent.Sync()
}
