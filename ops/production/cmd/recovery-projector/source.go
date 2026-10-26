package main

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func validateOptions(config options) error {
	seen := map[string]bool{}
	for _, value := range []string{config.source, config.api, config.worker} {
		if !filepath.IsAbs(value) || filepath.Clean(value) != value || seen[value] {
			return fmt.Errorf("directories must be distinct canonical absolute paths")
		}
		seen[value] = true
		info, err := os.Lstat(value)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("directory is fenced")
		}
	}
	return nil
}

func readSource(source string) ([]byte, []byte, error) {
	bundle, err := readRegular(
		filepath.Join(source, recoveryv1.BundleFilename),
		recoveryv1.MaximumBundleBytes,
	)
	if err != nil {
		return nil, nil, err
	}
	parsed, err := recoveryv1.Parse(bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("recovery bundle is invalid")
	}
	seed, err := readRegular(
		filepath.Join(source, recoveryv1.ControlSignerFilename),
		ed25519.SeedSize,
	)
	if err != nil || len(seed) != ed25519.SeedSize {
		clear(seed)
		return nil, nil, fmt.Errorf("recovery control signer is invalid")
	}
	public, found := parsed.PublicKey(recoveryv1.PurposeRecoveryControl)
	derived := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if !found || !bytes.Equal(public[:], derived) {
		clear(seed)
		return nil, nil, fmt.Errorf("recovery control signer does not match bundle")
	}
	return bundle, seed, nil
}

func readRegular(path string, maximum int) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() ||
		info.Size() < 1 || info.Size() > int64(maximum) {
		return nil, fmt.Errorf("recovery source artifact is invalid")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read recovery source artifact")
	}
	return content, nil
}
