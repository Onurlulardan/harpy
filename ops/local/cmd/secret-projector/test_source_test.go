package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func writeTestSource(t *testing.T, now time.Time) string {
	t.Helper()
	root := t.TempDir()
	names := make(map[string]bool)
	for _, target := range secretTargets() {
		for _, file := range target.files {
			names[file.name] = true
		}
	}
	names["unprojected-localconfig-artifact.txt"] = true
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	manifest := sourceManifest{
		Version:                  1,
		ArtifactContractRevision: supportedArtifactContractRevision,
		GenerationID:             base64.RawURLEncoding.EncodeToString(make([]byte, 16)),
		GeneratedAt:              now.Format(time.RFC3339),
		RenewAfter:               now.Add(23 * 24 * time.Hour).Format(time.RFC3339),
		ExpiresAt:                now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
	}
	for _, name := range sorted {
		content := []byte("fixture:" + name)
		digest := sha256.Sum256(content)
		manifest.Files = append(manifest.Files, sourceFile{
			Name: name, SHA256: hex.EncodeToString(digest[:]), Size: int64(len(content)),
		})
		if err := os.WriteFile(filepath.Join(root, name), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	content, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func noOwnershipChange(string, int, int) error { return nil }
