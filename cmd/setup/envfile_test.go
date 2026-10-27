package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWritePrivateReplacesBroadTempAndKeepsDestination(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "secret")
	if err := os.WriteFile(destination, []byte("previous"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := destination + ".tmp"
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(destination, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(stale, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writePrivate(destination, []byte("current")); err != nil {
		t.Fatal(err)
	}
	assertPrivateMode(t, destination)
	if body := mustRead(t, destination); body != "current" {
		t.Fatalf("destination = %q", body)
	}
	if body := mustRead(t, stale); body != "stale" {
		t.Fatalf("fixed temp name was reused: %q", body)
	}

	previous := renamePath
	t.Cleanup(func() { renamePath = previous })
	renamePath = failTempReplace
	if err := writePrivate(destination, []byte("lost")); err == nil {
		t.Fatal("expected replace failure")
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatalf("destination missing after failed replace: %v", err)
	}
	if body := mustRead(t, destination); body != "current" {
		t.Fatalf("destination after failed replace = %q", body)
	}
}

func failTempReplace(oldName, newName string) error {
	if strings.HasPrefix(filepath.Base(oldName), ".tmp-") {
		return errors.New("replace failed")
	}
	return os.Rename(oldName, newName)
}

func assertPrivateMode(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if runtime.GOOS == "windows" {
		if perm&0o200 == 0 {
			t.Fatalf("mode %o", perm)
		}
		return
	}
	if perm != 0o600 {
		t.Fatalf("mode %o", perm)
	}
}
