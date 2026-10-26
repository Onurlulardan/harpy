package devjournal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func PublishDevelopmentRecoveryTrust(
	authorityRoot string,
	recoveryRoot string,
) error {
	current := filepath.Join(authorityRoot, "current")
	if !validDevelopmentDirectory(current) || !validDevelopmentDirectory(recoveryRoot) {
		return fmt.Errorf("development recovery trust directories are invalid")
	}
	source, err := readDevelopmentRegular(
		filepath.Join(recoveryRoot, recoveryv1.BundleFilename),
		recoveryv1.MaximumBundleBytes,
	)
	if err != nil {
		return fmt.Errorf("read development recovery trust: %w", err)
	}
	defer clear(source)
	if _, err := recoveryv1.Parse(source); err != nil {
		return fmt.Errorf("verify development recovery trust: %w", err)
	}
	target := filepath.Join(current, recoveryv1.BundleFilename)
	existing, err := readDevelopmentRegular(target, recoveryv1.MaximumBundleBytes)
	if err == nil && bytes.Equal(existing, source) {
		clear(existing)
		return nil
	}
	clear(existing)
	if err != nil && !os.IsNotExist(err) {
		if _, statErr := os.Lstat(target); statErr == nil {
			return fmt.Errorf("development recovery trust target is invalid")
		}
	}
	temporary, err := os.CreateTemp(current, ".recovery-trust-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	published := false
	defer func() {
		_ = temporary.Close()
		if !published {
			_ = os.Remove(name)
		}
	}()
	if err = temporary.Chmod(0o400); err != nil {
		return err
	}
	if _, err = temporary.Write(source); err != nil {
		return err
	}
	if err = temporary.Sync(); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, target); err != nil {
		return fmt.Errorf("publish development recovery trust: %w", err)
	}
	published = true
	return nil
}
