package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func writePasswordBlocklists(dir string) (string, error) {
	commonBody := []byte("local-common")
	compromisedBody := []byte("local-compromised")
	common, err := writeRuntimeFile(dir, "password-common.bin", commonBody)
	if err != nil {
		return "", err
	}
	compromised, err := writeRuntimeFile(dir, "password-compromised.bin", compromisedBody)
	if err != nil {
		return "", err
	}
	document := map[string]any{
		"version": 1, "normalization": "unicode_nfc_v1", "digest": "sha256",
		"common": map[string]string{
			"path": common, "sha256": contentDigest(commonBody),
		},
		"compromised": map[string]string{
			"path": compromised, "sha256": contentDigest(compromisedBody),
		},
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	path, err := writeRuntimeFile(dir, "password-blocklists.json", encoded)
	if err != nil {
		return "", err
	}
	return runtimeLine("HARPY_IAM_PASSWORD_BLOCKLISTS_FILE", path), nil
}

func contentDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
