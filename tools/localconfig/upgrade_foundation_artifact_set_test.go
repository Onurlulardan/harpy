package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRevisionElevenAddsFoundationCredentialsWithoutRotatingRuntime(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	output := writeRevisionElevenArtifactSet(t, now)
	runtimeBefore, err := os.ReadFile(filepath.Join(output, "postgres-runtime-password.txt"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Generated ||
		result.Manifest.ArtifactContractRevision != currentArtifactContractRevision ||
		len(result.Manifest.Files) != expectedGeneratedArtifactCount {
		t.Fatalf("upgrade result is incomplete: %#v", result.Manifest)
	}
	runtimeAfter, err := os.ReadFile(filepath.Join(output, "postgres-runtime-password.txt"))
	if err != nil || string(runtimeAfter) != string(runtimeBefore) {
		t.Fatal("foundation credential upgrade rotated the runtime password")
	}
	content, err := os.ReadFile(filepath.Join(output, "initialize.env"))
	if err != nil || !strings.Contains(string(content), "HARPY_FOUNDATION_ADMIN_DSN_FILE=") ||
		strings.Contains(string(content), "postgres-migrator-dsn") {
		t.Fatalf("initialize.env = %s (%v)", content, err)
	}
}

func writeRevisionElevenArtifactSet(t *testing.T, now time.Time) string {
	t.Helper()
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	drop := make(map[string]bool, len(foundationCredentialFiles))
	for _, name := range foundationCredentialFiles {
		drop[name] = true
	}
	files := manifest.Files[:0]
	for _, file := range manifest.Files {
		if !drop[file.Name] {
			files = append(files, file)
		}
	}
	manifest.Files = files
	manifest.ArtifactContractRevision = foundationCredentialUpgradeSourceRevision
	output := filepath.Join(t.TempDir(), "secrets")
	if err := os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if err := os.WriteFile(
			filepath.Join(output, file.Name), set.files[file.Name], 0o600,
		); err != nil {
			t.Fatal(err)
		}
	}
	content, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, manifestFileName), content, 0o600); err != nil {
		t.Fatal(err)
	}
	return output
}
