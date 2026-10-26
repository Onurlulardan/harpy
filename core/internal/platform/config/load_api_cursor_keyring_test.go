package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAPIRequiresCursorKeyring(t *testing.T) {
	_, err := loadAPI(validEnvironment())
	if err == nil || !strings.Contains(err.Error(), cursorKeyringEnvironmentKey) {
		t.Fatalf("missing API cursor keyring was not rejected: %v", err)
	}
}

func TestLoadAPIRejectsBothCursorKeyringSources(t *testing.T) {
	entries := append(
		validAPIEnvironment(),
		cursorKeyringEnvironmentKey+"_FILE="+filepath.Join(t.TempDir(), "cursor.json"),
	)
	_, err := loadAPI(entries)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("cursor keyring source collision was not rejected: %v", err)
	}
}

func TestLoadAPIAcceptsCursorKeyringFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cursor.json")
	if err := os.WriteFile(path, []byte(validCursorKeyringJSON()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := withoutEnvironmentKey(validAPIEnvironment(), cursorKeyringEnvironmentKey)
	entries = append(entries, cursorKeyringEnvironmentKey+"_FILE="+path)
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	current, err := keyring.Current()
	if err != nil {
		t.Fatal(err)
	}
	kid, err := current.KID()
	if err != nil {
		t.Fatal(err)
	}
	if kid != "current-1" {
		t.Fatal("cursor keyring file was not exposed through API config")
	}
}

func withoutEnvironmentKey(entries []string, removed string) []string {
	filtered := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasPrefix(entry, removed+"=") {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func TestZeroAPIRejectsCursorKeyringAccess(t *testing.T) {
	var zero API
	if _, err := zero.RestQueryCursorKeyring(); err == nil {
		t.Fatal("zero API exposed an invalid cursor keyring")
	}
}

func TestWorkerRequiresCursorKeyring(t *testing.T) {
	entries := withoutEnvironmentKey(validWorkerEnvironment(), cursorKeyringEnvironmentKey)
	_, err := loadWorker(entries)
	if err == nil || !strings.Contains(err.Error(), cursorKeyringEnvironmentKey) {
		t.Fatalf("worker without cursor keyring was not rejected: %v", err)
	}
}

func TestWorkerExposesCursorKeyring(t *testing.T) {
	worker, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := worker.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	current, err := keyring.Current()
	if err != nil {
		t.Fatal(err)
	}
	kid, err := current.KID()
	if err != nil || kid != "current-1" {
		t.Fatalf("worker cursor key = %q, %v", kid, err)
	}
}
