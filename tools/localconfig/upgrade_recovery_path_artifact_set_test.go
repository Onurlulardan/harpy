package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRevisionSevenUpgradesRecoveryPathsWithoutRotatingSecrets(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	set.files["api.env"] = bytes.ReplaceAll(set.files["api.env"],
		[]byte("/run/recovery-signing/current/"),
		[]byte("/run/recovery-signing/"))
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = recoveryPathUpgradeSourceRevision
	output := filepath.Join(t.TempDir(), "secrets")
	if err = os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if err = os.WriteFile(filepath.Join(output, file.Name), set.files[file.Name], 0o600); err != nil {
			t.Fatal(err)
		}
	}
	manifestContent, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(output, manifestFileName), manifestContent, 0o600); err != nil {
		t.Fatal(err)
	}
	secretBefore, err := os.ReadFile(filepath.Join(output, "postgres-runtime-password.txt"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	secretAfter, err := os.ReadFile(filepath.Join(output, "postgres-runtime-password.txt"))
	if err != nil || !bytes.Equal(secretBefore, secretAfter) {
		t.Fatal("recovery path upgrade rotated the runtime database secret")
	}
	api, err := os.ReadFile(filepath.Join(output, "api.env"))
	if err != nil || !bytes.Contains(api,
		[]byte("HARPY_RECOVERY_TRUST_BUNDLE_PATH=/run/recovery-signing/current/")) {
		t.Fatal("recovery path upgrade did not refresh api.env")
	}
	if !result.Generated || result.Manifest.ArtifactContractRevision !=
		currentArtifactContractRevision {
		t.Fatalf("recovery path upgrade result = %#v", result)
	}
}
