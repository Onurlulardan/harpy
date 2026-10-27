package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretFormattingAlwaysRedacts(t *testing.T) {
	secret := Secret{value: "must-not-appear"}
	for _, rendered := range []string{
		fmt.Sprint(secret), fmt.Sprintf("%s", secret), fmt.Sprintf("%v", secret),
		fmt.Sprintf("%+v", secret), fmt.Sprintf("%#v", secret), fmt.Sprintf("%q", secret),
	} {
		if strings.Contains(rendered, secret.Value()) || rendered != "[REDACTED]" {
			t.Fatalf("secret formatter leaked or changed marker: %q", rendered)
		}
	}
}

func TestSecretValueAndFileAreMutuallyExclusive(t *testing.T) {
	entries := append(validAPIEnvironment(), "HARPY_POSTGRES_DSN_FILE=C:\\secret")
	_, err := loadAPI(entries)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected XOR failure, got %v", err)
	}
}

func TestSecretFileResolvesSymlinkAndOneNewline(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	link := filepath.Join(directory, "projected-secret")
	if err := os.WriteFile(target, []byte("postgresql://user@example.invalid/db?sslmode=disable\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	entries := []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN_FILE=" + link,
		cursorKeyringEnvironmentEntry(),
	}
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatal(err)
	}
	value := loaded.LongRunning().Runtime().Postgres().DSN().Value()
	if value != "postgresql://user@example.invalid/db?sslmode=disable" {
		t.Fatalf("unexpected loaded secret: %q", value)
	}
}

func TestSecretFileRejectsNULAndDoesNotLeakPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sensitive-location")
	if err := os.WriteFile(path, []byte("invalid\x00value"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN_FILE=" + path,
		cursorKeyringEnvironmentEntry(),
	}
	_, err := loadAPI(entries)
	if err == nil || !strings.Contains(err.Error(), "prohibited byte") {
		t.Fatalf("expected NUL rejection, got %v", err)
	}
	if strings.Contains(err.Error(), path) || strings.Contains(err.Error(), "invalid") {
		t.Fatalf("secret error leaked path or value: %v", err)
	}
}

func TestSecretFileRejectsRelativeAndDirectoryPaths(t *testing.T) {
	for _, path := range []string{"relative-secret", t.TempDir()} {
		entries := []string{
			"HARPY_ENVIRONMENT=development",
			"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
			"HARPY_POSTGRES_DSN_FILE=" + path,
			cursorKeyringEnvironmentEntry(),
		}
		if _, err := loadAPI(entries); err == nil {
			t.Fatalf("expected secret path rejection for %q", path)
		}
	}
}
