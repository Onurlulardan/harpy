package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSecretFileAcceptsOwnerOnlyModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix mode bits are enforced off Windows; Ubuntu CI is the proof")
	}
	const secret = "postgresql://user@example.invalid/db?sslmode=disable"
	for _, mode := range []os.FileMode{0o600, 0o400} {
		t.Run(mode.String(), func(t *testing.T) {
			path := writeModeFile(t, secret, mode)
			loaded, err := loadAPI(secretFileEnvironment(path))
			if err != nil {
				t.Fatal(err)
			}
			if loaded.LongRunning().Runtime().Postgres().DSN().Value() != secret {
				t.Fatalf("mode %o rejected an owner-only secret", mode)
			}
		})
	}
}

func TestSecretFileRejectsGroupAndOtherPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix mode bits are enforced off Windows; Ubuntu CI is the proof")
	}
	for _, mode := range []os.FileMode{0o640, 0o644, 0o444} {
		t.Run(mode.String(), func(t *testing.T) {
			path := writeModeFile(t, "postgresql://user@example.invalid/db?sslmode=disable", mode)
			_, err := loadAPI(secretFileEnvironment(path))
			if err == nil || !strings.Contains(err.Error(), "permissions are too broad") {
				t.Fatalf("mode %o: %v", mode, err)
			}
			if strings.Contains(err.Error(), path) {
				t.Fatalf("error leaked path: %v", err)
			}
		})
	}
}

func writeModeFile(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != mode {
		t.Fatalf("filesystem mode = %o, want %o", info.Mode().Perm(), mode)
	}
	return path
}
