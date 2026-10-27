package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsumeLoadsProfileKeysAndSkipsOtherProfiles(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".env")
	body := strings.Join([]string{
		"# local",
		"",
		"HARPY_ENVIRONMENT=development",
		"HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:8080",
		"GOMEMLIMIT=512MiB",
		"POSTGRES_USER=harpy",
	}, "\n")
	if err := os.WriteFile(path, []byte(body+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HARPY_ENVIRONMENT", "")
	t.Setenv("HARPY_PUBLIC_LISTEN_ADDR", "")
	t.Setenv("GOMEMLIMIT", "")
	t.Setenv("POSTGRES_USER", "")
	remaining, err := Consume([]string{"--env-file", path, "deploy", "up"}, "worker")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(remaining, " ") != "deploy up" {
		t.Fatalf("remaining = %v", remaining)
	}
	if os.Getenv("HARPY_ENVIRONMENT") != "development" || os.Getenv("GOMEMLIMIT") != "512MiB" {
		t.Fatalf("worker keys missing: %s %s", os.Getenv("HARPY_ENVIRONMENT"), os.Getenv("GOMEMLIMIT"))
	}
	if os.Getenv("HARPY_PUBLIC_LISTEN_ADDR") != "" {
		t.Fatal("worker applied an API listen address")
	}
	if os.Getenv("POSTGRES_USER") != "harpy" {
		t.Fatal("compose key was not loaded")
	}
}

func TestConsumeRejectsUnknownHarpyKeyByName(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".env")
	if err := os.WriteFile(path, []byte("HARPY_TYPO=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Consume([]string{"--env-file=" + path}, "api")
	if err == nil || !strings.Contains(err.Error(), "HARPY_TYPO") || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("error = %v", err)
	}
}

func TestConsumeNamesMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.env")
	_, err := Consume([]string{"--env-file", missing}, "api")
	if err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("error = %v", err)
	}
}
