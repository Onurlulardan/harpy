package configfixture

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

type passwordBlocklistCoordinate struct {
	path   string
	digest string
}

type ArtifactFixture interface {
	Helper()
	TempDir() string
	Fatalf(string, ...any)
}

type IAMAPISetter interface {
	EnvironmentSetter
	ArtifactFixture
}

func passwordBlocklistFixture(directory string) (string, error) {
	common, err := writePasswordBlocklistFixture(
		directory, "common-v1.bin", blocklistv1.KindCommon,
		[]byte("fixture common password"),
	)
	if err != nil {
		return "", err
	}
	compromised, err := writePasswordBlocklistFixture(
		directory, "compromised-v1.bin", blocklistv1.KindCompromised,
		[]byte("fixture compromised password"),
	)
	if err != nil {
		return "", err
	}
	return marshal(map[string]any{
		"version": 1, "normalization": "unicode_nfc_v1", "digest": "sha256",
		"common": map[string]string{
			"path": common.path, "sha256": common.digest,
		},
		"compromised": map[string]string{
			"path": compromised.path, "sha256": compromised.digest,
		},
	}), nil
}

func writePasswordBlocklistFixture(
	directory string,
	name string,
	kind blocklistv1.Kind,
	candidate []byte,
) (passwordBlocklistCoordinate, error) {
	record := sha256.Sum256(candidate)
	content, err := blocklistv1.Encode(kind, [][32]byte{record})
	if err != nil {
		return passwordBlocklistCoordinate{}, err
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		clear(content)
		return passwordBlocklistCoordinate{}, fmt.Errorf("write fixture: %w", err)
	}
	digest := sha256.Sum256(content)
	clear(content)
	return passwordBlocklistCoordinate{path: path, digest: hex.EncodeToString(digest[:])}, nil
}
