package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureArtifactSetIsIdempotentAndFailsClosedOnIntegrityLoss(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := filepath.Join(t.TempDir(), "secrets")
	first, err := ensureArtifactSet(output, now)
	if err != nil || !first.Generated {
		t.Fatalf("first generation: generated=%t err=%v", first.Generated, err)
	}
	firstManifest, err := os.ReadFile(filepath.Join(output, manifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	second, err := ensureArtifactSet(output, now.Add(time.Hour))
	if err != nil || second.Generated || second.Manifest.GenerationID != first.Manifest.GenerationID {
		t.Fatalf("idempotent generation differs: generated=%t err=%v", second.Generated, err)
	}
	secondManifest, err := os.ReadFile(filepath.Join(output, manifestFileName))
	if err != nil || !bytes.Equal(firstManifest, secondManifest) {
		t.Fatalf("current set changed: %v", err)
	}
	if err := os.Remove(filepath.Join(output, "smtp-password.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureArtifactSet(output, now.Add(2*time.Hour)); err == nil {
		t.Fatal("missing artifact was silently regenerated")
	}
	if removed, _, err := resetArtifactSet(output, now.Add(2*time.Hour)); err == nil || removed {
		t.Fatalf("integrity-damaged set was reset: removed=%t err=%v", removed, err)
	}
}

func TestExpiredArtifactSetRequiresExplicitVerifiedReset(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	parent := t.TempDir()
	output := filepath.Join(parent, "secrets")
	first, err := ensureArtifactSet(output, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ensureArtifactSet(output, now.Add(24*24*time.Hour)); err == nil {
		t.Fatal("expiring set was silently rotated")
	}
	manifest, err := os.ReadFile(filepath.Join(output, manifestFileName))
	if err != nil || !bytes.Contains(manifest, []byte(first.Manifest.GenerationID)) {
		t.Fatalf("expired generation changed before reset: %v", err)
	}
	removed, _, err := resetArtifactSet(output, now.Add(24*24*time.Hour))
	if err != nil || !removed {
		t.Fatalf("verified expired set was not reset: removed=%t err=%v", removed, err)
	}
	if _, err := os.Lstat(output); !os.IsNotExist(err) {
		t.Fatalf("reset target still exists: %v", err)
	}
	renewed, err := ensureArtifactSet(output, now.Add(24*24*time.Hour))
	if err != nil || !renewed.Generated ||
		renewed.Manifest.GenerationID == first.Manifest.GenerationID {
		t.Fatalf("fresh set was not generated after reset: generated=%t err=%v",
			renewed.Generated, err)
	}
}
