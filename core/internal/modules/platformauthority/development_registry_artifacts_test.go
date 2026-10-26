package platformauthority

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDevelopmentRegistryArtifactsRecoverInterruptedPublication(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "current")
	if err := os.Mkdir(current, 0o700); err != nil {
		t.Fatal(err)
	}
	ackSeed := filepath.Join(root, "ack.seed")
	releaseSeed := filepath.Join(root, "release.seed")
	writeDevelopmentSeed(t, ackSeed, 11)
	writeDevelopmentSeed(t, releaseSeed, 29)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	if err := ReconcileDevelopmentRegistryArtifacts(
		root, ackSeed, releaseSeed, now,
	); err != nil {
		t.Fatal(err)
	}
	trustPath := filepath.Join(current, DevelopmentAckTrustFilename)
	first, err := os.ReadFile(trustPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(trustPath); err != nil {
		t.Fatal(err)
	}
	if err = ReconcileDevelopmentRegistryArtifacts(
		root, ackSeed, releaseSeed, now.Add(time.Minute),
	); err != nil {
		t.Fatal(err)
	}
	recovered, err := os.ReadFile(trustPath)
	if err != nil || len(recovered) == 0 || string(recovered) == string(first) {
		t.Fatalf("recovered trust artifact is invalid: %v", err)
	}
	if err = ReconcileDevelopmentRegistryArtifacts(
		root, ackSeed, releaseSeed, now.Add(2*time.Minute),
	); err != nil {
		t.Fatal(err)
	}
}

func TestDevelopmentRegistryRejectsAckReleaseKeyReuse(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "current"), 0o700); err != nil {
		t.Fatal(err)
	}
	seed := filepath.Join(root, "shared.seed")
	writeDevelopmentSeed(t, seed, 17)
	if err := ReconcileDevelopmentRegistryArtifacts(
		root, seed, seed, time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	); err == nil {
		t.Fatal("acknowledgement and release key reuse was accepted")
	}
}

func writeDevelopmentSeed(t *testing.T, path string, marker byte) {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = marker + byte(index)
	}
	if err := os.WriteFile(path, seed, 0o400); err != nil {
		t.Fatal(err)
	}
}
