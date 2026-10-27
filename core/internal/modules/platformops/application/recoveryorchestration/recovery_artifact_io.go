package recoveryorchestration

import (
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func readRecoveryArtifact(
	path string,
	maximum int64,
	sensitive bool,
) ([]byte, error) {
	return readRecoveryFile(path, maximum, sensitive)
}

func readRecoveryFile(path string, maximum int64, sensitive bool) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > maximum ||
		sensitive && runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("recovery artifact is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recovery artifact is unavailable")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("recovery artifact changed while opening")
	}
	document, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(document)) > maximum {
		clear(document)
		return nil, fmt.Errorf("recovery artifact exceeds its limit")
	}
	return document, nil
}

func readRecoverySigningKey(path string) (ed25519.PrivateKey, error) {
	const seedSize = 32
	seed, err := readRecoveryArtifact(path, seedSize, true)
	if err != nil || len(seed) != seedSize {
		clear(seed)
		return nil, fmt.Errorf("recovery signing key is invalid")
	}
	private := ed25519.NewKeyFromSeed(seed)
	clear(seed)
	return private, nil
}

func recoveryProviderPath(kind, handle string) (string, error) {
	if filepath.IsAbs(handle) {
		return filepath.Clean(handle), nil
	}
	if kind != "local_keyring" ||
		filepath.Base(handle) != handle {
		return "", fmt.Errorf("recovery key provider handle is invalid")
	}
	return filepath.Join("/run/recovery-state/providers", handle+".json"), nil
}
