package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOwnedOutputPublishesAtomicallyAndRejectsReplacement(t *testing.T) {
	target := filepath.Join(t.TempDir(), "release")
	first, err := beginOwnedOutput(target, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(first.staging, "value"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := first.commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := beginOwnedOutput(target, false); err == nil {
		t.Fatal("owned output replacement was accepted")
	}
	value, err := os.ReadFile(filepath.Join(target, "value"))
	if err != nil || string(value) != "one" {
		t.Fatalf("published value = %q err=%v", value, err)
	}
}

func TestOwnedOutputRejectsUnmarkedDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "release")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := beginOwnedOutput(target, true); err == nil {
		t.Fatal("unmarked output directory accepted")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("unmarked output directory was changed: %v", err)
	}
}

func TestOwnedOutputReplacementIsMarkerBoundAndAtomic(t *testing.T) {
	target := filepath.Join(t.TempDir(), "release")
	first, err := beginOwnedOutput(target, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(first.staging, "value"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := first.commit(); err != nil {
		t.Fatal(err)
	}
	replacement, err := beginOwnedOutput(target, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(replacement.staging, "value"), []byte("two"), 0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := replacement.commit(); err != nil {
		t.Fatal(err)
	}
	value, err := os.ReadFile(filepath.Join(target, "value"))
	if err != nil || string(value) != "two" {
		t.Fatalf("replacement value = %q err=%v", value, err)
	}
}
