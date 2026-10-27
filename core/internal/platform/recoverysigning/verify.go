package recoverysigning

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"runtime"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func verify(directory, environmentID string, reused bool) (Result, error) {
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 4 || !safeDirectoryMode(directory) {
		return Result{}, ErrFenced
	}
	expected := map[string]bool{recoveryv1.BundleFilename: true}
	for _, purpose := range []recoveryv1.Purpose{
		recoveryv1.PurposeRecoveryManifest,
		recoveryv1.PurposeRecoveryControl,
		recoveryv1.PurposeRestoreReceipt,
	} {
		expected[recoveryv1.Filename(purpose)] = true
	}
	for _, entry := range entries {
		if !expected[entry.Name()] || entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return Result{}, ErrFenced
		}
	}
	bundleBytes, err := readRegular(filepath.Join(directory, recoveryv1.BundleFilename),
		recoveryv1.MaximumBundleBytes, false)
	if err != nil {
		return Result{}, ErrFenced
	}
	bundle, err := recoveryv1.Parse(bundleBytes)
	if err != nil || bundle.EnvironmentID() != environmentID ||
		bundle.Generation() != recoveryv1.Generation {
		return Result{}, ErrFenced
	}
	for _, purpose := range []recoveryv1.Purpose{
		recoveryv1.PurposeRecoveryManifest,
		recoveryv1.PurposeRecoveryControl,
		recoveryv1.PurposeRestoreReceipt,
	} {
		seed, readErr := readRegular(filepath.Join(directory, recoveryv1.Filename(purpose)),
			ed25519.SeedSize, true)
		if readErr != nil || len(seed) != ed25519.SeedSize {
			return Result{}, ErrFenced
		}
		derived := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
		public, found := bundle.PublicKey(purpose)
		if !found || string(derived) != string(public[:]) {
			return Result{}, ErrFenced
		}
	}
	return Result{Generation: bundle.Generation(),
		TrustBundleSHA256: bundle.SHA256(), Reused: reused}, nil
}

func readRegular(path string, maximum int, sensitive bool) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > int64(maximum) ||
		runtime.GOOS != "windows" && sensitive && info.Mode().Perm()&0o077 != 0 {
		return nil, ErrFenced
	}
	return os.ReadFile(path)
}

func safeDirectoryMode(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir() &&
		(runtime.GOOS == "windows" || info.Mode().Perm()&0o077 == 0)
}
