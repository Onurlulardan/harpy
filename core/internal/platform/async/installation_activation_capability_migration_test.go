package async

import (
	"os"
	"path/filepath"
	"testing"
)

func readInstallationActivationMigration(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(
		"migrations", "fresh", name+".fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	return body
}
