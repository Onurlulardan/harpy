package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRevisionFourUpgradesWithoutRotatingExistingSecrets(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	output := writeRevisionFourArtifactSet(t, now)
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
		t.Fatal("recovery upgrade rotated an existing runtime secret")
	}
	if _, err := os.Stat(filepath.Join(output, "recovery-write-credential.json")); err != nil {
		t.Fatal("recovery upgrade did not add purpose credentials")
	}
	if _, current, err := inspectArtifactSet(output, now.Add(time.Minute), false); err != nil ||
		!current {
		t.Fatalf("upgraded artifact set is not current: %v", err)
	}
}

func TestRevisionFourUpgradeRejectsTamperedSource(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	output := writeRevisionFourArtifactSet(t, now)
	if err := os.WriteFile(filepath.Join(output, "postgres-runtime-password.txt"),
		[]byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureArtifactSet(output, now.Add(time.Minute)); err == nil {
		t.Fatal("tampered revision-four source was upgraded")
	}
	staged, backup := upgradePaths(output)
	for _, path := range []string{staged, backup} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("failed upgrade left swap state at %s", path)
		}
	}
}

func TestUpgradeRecoveryCleansCompletedSwapBackup(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	output := filepath.Join(t.TempDir(), "secrets")
	if _, err := ensureArtifactSet(output, now); err != nil {
		t.Fatal(err)
	}
	_, backup := upgradePaths(output)
	if err := os.Mkdir(backup, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := recoverArtifactUpgrade(output, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(backup); !os.IsNotExist(err) {
		t.Fatal("completed upgrade backup was not removed")
	}
}

func writeRevisionFourArtifactSet(t *testing.T, now time.Time) string {
	t.Helper()
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	recoveryNames := make(map[string]bool, len(recoveryArtifactNames))
	for _, name := range recoveryArtifactNames {
		recoveryNames[name] = true
	}
	for _, name := range authorityArtifactNames {
		recoveryNames[name] = true
	}
	recoveryNames["tenant-migration-executor.env"] = true
	files := manifest.Files[:0]
	for _, file := range manifest.Files {
		if !recoveryNames[file.Name] {
			files = append(files, file)
		}
	}
	manifest.Files = files
	manifest.ArtifactContractRevision = recoveryUpgradeSourceRevision
	output := filepath.Join(t.TempDir(), "secrets")
	if err := os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if err := os.WriteFile(filepath.Join(output, file.Name), set.files[file.Name], 0o600); err != nil {
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
