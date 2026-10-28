package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectTargetCopiesOnlyAllowlistAndIsIdempotent(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	sourceRoot := writeTestSource(t, now)
	source, err := verifySource(sourceRoot, now)
	if err != nil {
		t.Fatal(err)
	}
	target := runtimeTarget("test", "local-ca.crt", "postgres-runtime-dsn.txt")
	targetRoot := t.TempDir()
	keepTempDirRemovable(t, targetRoot)
	if err := projectTarget(targetRoot, source, target, noOwnershipChange); err != nil {
		t.Fatal(err)
	}
	if err := projectTarget(targetRoot, source, target, noOwnershipChange); err != nil {
		t.Fatalf("repeat projection failed: %v", err)
	}
	entries, err := os.ReadDir(targetRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("projected %d entries, want 3", len(entries))
	}
	if _, err := os.Stat(filepath.Join(targetRoot, "smtp-password.txt")); !os.IsNotExist(err) {
		t.Fatal("non-allowlisted worker secret leaked into target")
	}
}

func TestProjectTargetRejectsChangedProjection(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	sourceRoot := writeTestSource(t, now)
	source, err := verifySource(sourceRoot, now)
	if err != nil {
		t.Fatal(err)
	}
	target := runtimeTarget("test", "local-ca.crt")
	targetRoot := t.TempDir()
	keepTempDirRemovable(t, targetRoot)
	if err := projectTarget(targetRoot, source, target, noOwnershipChange); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(targetRoot, "local-ca.crt")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := projectTarget(targetRoot, source, target, noOwnershipChange); err == nil {
		t.Fatal("changed projection was overwritten")
	}
}

func TestProjectTargetRecoversOwnedPendingWrite(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	sourceRoot := writeTestSource(t, now)
	source, err := verifySource(sourceRoot, now)
	if err != nil {
		t.Fatal(err)
	}
	target := runtimeTarget("test", "local-ca.crt", "postgres-runtime-dsn.txt")
	targetRoot := t.TempDir()
	keepTempDirRemovable(t, targetRoot)
	marker, err := marshalMarker(markerFor(source, target))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, pendingMarker), marker, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, "local-ca.crt"), []byte("partial"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := projectTarget(targetRoot, source, target, noOwnershipChange); err != nil {
		t.Fatal(err)
	}
}

func keepTempDirRemovable(t *testing.T, dir string) {
	t.Helper()
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
}
