package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func validIAMPasswordBlocklistsJSON() string {
	directory := filepath.Join(os.TempDir(), "harpy-password-blocklists")
	document := map[string]any{
		"version": 1, "normalization": "unicode_nfc_v1", "digest": "sha256",
		"common": map[string]string{
			"path":   filepath.Join(directory, "common-v1.bin"),
			"sha256": strings.Repeat("1", 64),
		},
		"compromised": map[string]string{
			"path":   filepath.Join(directory, "compromised-v1.bin"),
			"sha256": strings.Repeat("2", 64),
		},
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}
