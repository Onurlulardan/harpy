package postgres

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestInstallationCredentialDeliveryMigrationsArePairedAndSecretFree(t *testing.T) {
	for version := 20260722142000; version <= 20260722142010; version++ {
		prefix := strconv.Itoa(version) + "_"
		fresh := migrationWithPrefix(t, "fresh", prefix)
		adopted := migrationWithPrefix(t, "v1_adopted", prefix)
		if strings.Replace(fresh, ".fresh.sql", "", 1) !=
			strings.Replace(adopted, ".v1_adopted.sql", "", 1) {
			t.Fatalf("installation delivery pair differs for %d", version)
		}
		document, err := os.ReadFile(filepath.Join(platformMigrationDirectory("fresh"), fresh))
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(document))
		for _, forbidden := range []string{"plaintext", "password_hash", "secret_path", "file_path"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%s persists forbidden field %q", fresh, forbidden)
			}
		}
	}
}

func migrationWithPrefix(t *testing.T, origin, prefix string) string {
	t.Helper()
	entries, err := os.ReadDir(platformMigrationDirectory(origin))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			return entry.Name()
		}
	}
	t.Fatalf("missing migration %s/%s", origin, prefix)
	return ""
}
