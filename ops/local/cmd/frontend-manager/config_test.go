package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalAPIPortReadsLocalEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.local")
	if err := os.WriteFile(path, []byte("IGNORED=x\nLOCAL_API_PORT=8123\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	port, err := localAPIPort(path)
	if err != nil {
		t.Fatal(err)
	}
	if port != "8123" {
		t.Fatalf("port = %q, want 8123", port)
	}
}

func TestLocalAPIPortDefaultsWhenFileIsAbsent(t *testing.T) {
	port, err := localAPIPort(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if port != "8000" {
		t.Fatalf("port = %q, want 8000", port)
	}
}

func TestSetEnvironmentReplacesExistingValue(t *testing.T) {
	got := setEnvironment([]string{"A=1", "VITE_API_BASE_URL=old"}, "VITE_API_BASE_URL", "")
	if len(got) != 2 || got[1] != "VITE_API_BASE_URL=" {
		t.Fatalf("environment = %#v", got)
	}
}

func TestLocalFrontendReadyURLUsesCanonicalOrigin(t *testing.T) {
	if localFrontendReadyURL != "http://localhost:5173/" {
		t.Fatalf("ready URL = %q", localFrontendReadyURL)
	}
}
