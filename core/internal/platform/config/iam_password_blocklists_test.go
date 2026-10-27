package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIAMPasswordBlocklistsLoadExactCoordinates(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	loaded, err := loadAPIAt(validAPIEnvironmentAt(now), now)
	if err != nil {
		t.Fatal(err)
	}
	iamSettings, err := loaded.IAM()
	if err != nil {
		t.Fatal(err)
	}
	blocklists, err := iamSettings.PasswordBlocklists()
	if err != nil {
		t.Fatal(err)
	}
	common, err := blocklists.Common()
	if err != nil {
		t.Fatal(err)
	}
	path, _ := common.Path()
	digest, _ := common.SHA256()
	if !strings.HasSuffix(path.Value(), filepath.Join("harpy-password-blocklists", "common-v1.bin")) ||
		strings.Repeat("1", 64) != formatSHA256(digest) {
		t.Fatal("common artifact coordinate changed during config parsing")
	}
}

func TestIAMPasswordBlocklistsRejectStructuralAndPurposeDrift(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	base := passwordBlocklistFixtureDocument(t)
	tests := map[string]func(map[string]any){
		"unknown member": func(value map[string]any) { value["extra"] = true },
		"normalization":  func(value map[string]any) { value["normalization"] = "NFKC" },
		"uppercase digest": func(value map[string]any) {
			value["common"].(map[string]any)["sha256"] = strings.Repeat("A", 64)
		},
		"shared path": func(value map[string]any) {
			value["compromised"].(map[string]any)["path"] = value["common"].(map[string]any)["path"]
		},
		"shared digest": func(value map[string]any) {
			value["compromised"].(map[string]any)["sha256"] = value["common"].(map[string]any)["sha256"]
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := clonePasswordBlocklistFixture(t, base)
			mutate(value)
			document, _ := json.Marshal(value)
			entries := withEnvironment(validAPIEnvironmentAt(now),
				iamPasswordBlocklistsEnvironmentKey+"="+string(document))
			if _, err := loadAPIAt(entries, now); err == nil || strings.Contains(err.Error(), string(document)) {
				t.Fatalf("invalid or leaked blocklist config: %v", err)
			}
		})
	}
}

func TestIAMPasswordBlocklistsSupportOneRedactedFileSource(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	document := validIAMPasswordBlocklistsJSON()
	path := filepath.Join(t.TempDir(), "password-blocklists.json")
	if err := os.WriteFile(path, []byte(document+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := withoutEnvironmentKey(validAPIEnvironmentAt(now), iamPasswordBlocklistsEnvironmentKey)
	entries = append(entries, iamPasswordBlocklistsEnvironmentKey+"_FILE="+path)
	if _, err := loadAPIAt(entries, now); err != nil {
		t.Fatalf("file-only password blocklist config rejected: %v", err)
	}
	both := append(validAPIEnvironmentAt(now), iamPasswordBlocklistsEnvironmentKey+"_FILE="+path)
	if _, err := loadAPIAt(both, now); err == nil || strings.Contains(err.Error(), document) {
		t.Fatalf("source collision was accepted or leaked config: %v", err)
	}
}

func passwordBlocklistFixtureDocument(t *testing.T) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(validIAMPasswordBlocklistsJSON()), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func clonePasswordBlocklistFixture(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	document, _ := json.Marshal(value)
	var cloned map[string]any
	if err := json.Unmarshal(document, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func formatSHA256(value [32]byte) string {
	const alphabet = "0123456789abcdef"
	encoded := make([]byte, 64)
	for index, current := range value {
		encoded[index*2] = alphabet[current>>4]
		encoded[index*2+1] = alphabet[current&15]
	}
	return string(encoded)
}
