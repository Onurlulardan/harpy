package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestExactInventoryAllowsOnlyTheDeclaredSource(t *testing.T) {
	root := t.TempDir()
	secretPath := filepath.Join(root, "initial-admin-password")
	writeInventoryFixture(t, secretPath, testSecret, 0o600)
	writeInventoryFixture(t, filepath.Join(root, "receipt.json"), `{"digest":"safe"}`, 0o600)
	arguments := []string{
		"-needle-file", secretPath, "-allow-file", secretPath, "-path", root,
		"-stdin-label", "logs",
	}
	verified, err := runExact(context.Background(), arguments, strings.NewReader("safe logs"))
	if err != nil || verified < 3 {
		t.Fatalf("verified=%d error=%v", verified, err)
	}
	writeInventoryFixture(t, filepath.Join(root, "job.json"), "payload="+testSecret, 0o600)
	if _, err = runExact(context.Background(), arguments, strings.NewReader("safe logs")); err == nil {
		t.Fatal("file plaintext leak was accepted")
	}
}

func TestExactInventoryRejectsStreamBoundaryLeak(t *testing.T) {
	reader := bytes.NewBufferString(strings.Repeat("x", 64*1024-8) + testSecret)
	found, err := streamContains(context.Background(), reader, []byte(testSecret), 128*1024)
	if err != nil || !found {
		t.Fatalf("boundary leak found=%v error=%v", found, err)
	}
}

func TestExactInventoryRejectsEnvironmentLeak(t *testing.T) {
	t.Setenv("HARPY_INVENTORY_TEST", testSecret)
	if err := scanEnvironment([]byte(testSecret)); err == nil ||
		!strings.Contains(err.Error(), "HARPY_INVENTORY_TEST") ||
		strings.Contains(err.Error(), testSecret) {
		t.Fatalf("environment error = %v", err)
	}
}

func writeInventoryFixture(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}
