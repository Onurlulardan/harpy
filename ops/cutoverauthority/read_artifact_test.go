package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReadSecretStripsOneLineEndingAndRejectsNUL(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "secret")
	if err := os.WriteFile(path, []byte("value\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := readSecret(path)
	if err != nil || string(value) != "value" {
		t.Fatalf("secret=%q err=%v", value, err)
	}
	if err := os.WriteFile(path, []byte{'a', 0, 'b'}, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readSecret(path); err == nil {
		t.Fatal("NUL secret accepted")
	}
}

func TestReadArtifactRejectsOversizeAfterOpening(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact")
	if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, 5), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readArtifact(path, 4); err == nil {
		t.Fatal("oversize artifact accepted")
	}
}

func TestReadOpenedArtifactIsStableAcrossPathReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows open-handle rename semantics are covered by the ACL boundary")
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "artifact")
	if err := os.WriteFile(path, []byte("trusted"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := os.Rename(path, filepath.Join(directory, "original")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	document, err := readOpenedArtifact(file, maximumArtifactBytes, false)
	if err != nil || string(document) != "trusted" {
		t.Fatalf("opened handle changed across replacement: %q err=%v", document, err)
	}
}

func TestReadSecretRejectsGroupOrWorldPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows secret ACL enforcement is an operator boundary")
	}
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := readSecret(path); err == nil {
		t.Fatal("group-readable secret accepted")
	}
}
